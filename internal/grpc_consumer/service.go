package grpc_consumer

import (
	"context"
	"errors"
	"fmt"
	"go_video_streamer/internal/InputStreamShard"
	"go_video_streamer/internal/video_streaming"
	"log/slog"
	"slices"
	"sync"
)

type StreamQ struct {
	q              chan *InputStreamShard.StreamShard
	mu             sync.Mutex
	captureContext *video_streaming.CaptureContext
}

type GRPCVideoRCVServer struct {
	VideoRCVServer
	channelMapping   map[string]*StreamQ
	manuallyRemoved  []string
	editStreamsMutex sync.Mutex
}

func NewGRPCVideoRCVServer() *GRPCVideoRCVServer {
	return &GRPCVideoRCVServer{
		channelMapping:   make(map[string]*StreamQ),
		manuallyRemoved:  make([]string, 0),
		editStreamsMutex: sync.Mutex{},
	}
}

func (gs *GRPCVideoRCVServer) AddChannel(
	name string, ignoreFailover bool, streamParams *NewStream,
) (*StreamQ, error) {
	//
	gs.editStreamsMutex.Lock()
	defer gs.editStreamsMutex.Unlock()
	if slices.Contains(gs.manuallyRemoved, name) {
		if ignoreFailover {
			idx := slices.Index(gs.manuallyRemoved, name)
			gs.manuallyRemoved = append(gs.manuallyRemoved[:idx], gs.manuallyRemoved[idx+1:]...)
		} else {
			return nil, fmt.Errorf("channel already removed")
		}
	}
	channel := make(chan *InputStreamShard.StreamShard, 50)
	capture, err := NewGRPCCapture(&ConsumerConfig{
		grpcServer: gs,
	}, channel)
	if err != nil {
		panic(err)
	}
	streamWrapper := StreamQ{
		q:  channel,
		mu: sync.Mutex{},
		captureContext: video_streaming.CaptureContextFromCapture(
			capture,
			video_streaming.CaptureParams{
				FPS:    float64(streamParams.Fps),
				Width:  int(streamParams.Width),
				Height: int(streamParams.Height),
			},
			name,
		),
	}

	switch streamParams.StreamType {
	case 0:
		go video_streaming.ListenStreamToHLS(streamWrapper.captureContext, name)
		break

	default:
		panic(errors.New("invalid stream type"))
	}

	gs.channelMapping[name] = &streamWrapper
	return &streamWrapper, nil
}

func (gs *GRPCVideoRCVServer) RemoveChannel(name string) {
	gs.editStreamsMutex.Lock()
	defer gs.editStreamsMutex.Unlock()
	stream, ok := gs.channelMapping[name]
	if !ok {
		return
	}
	_, cancel := context.WithCancel(stream.captureContext.Context)
	cancel()
	delete(gs.channelMapping, name)
	if !slices.Contains(gs.manuallyRemoved, name) {
		gs.manuallyRemoved = append(gs.manuallyRemoved, name)
	}
}

func (gs *GRPCVideoRCVServer) RMStream(
	_ context.Context, stream *RemoveStream,
) (*EditStreamOperationResponse, error) {

	slog.Info(fmt.Sprintf("RMStream initialized via gRPC: name=%s", stream.Name))
	gs.RemoveChannel(stream.Name)
	return &EditStreamOperationResponse{Status: 200, Message: "OK"}, nil
}

func (gs *GRPCVideoRCVServer) AddStream(
	_ context.Context, stream *NewStream,
) (*EditStreamOperationResponse, error) {

	slog.Info(
		fmt.Sprintf(
			"New stream initialized via gRPC: name=%s, fps=%f, w=%d, h=%d",
			stream.Name, stream.Fps, stream.Width, stream.Height,
		),
	)

	_, err := gs.AddChannel(stream.Name, false, stream)
	if err == nil {
		return &EditStreamOperationResponse{Status: 200, Message: "OK"}, nil
	} else {
		return &EditStreamOperationResponse{Status: 500, Message: err.Error()}, nil
	}
}

func (gs *GRPCVideoRCVServer) StreamFrames(_ context.Context, frame *NamedFrame) (*VideoStreamResponse, error) {
	slog.Info(fmt.Sprintf("New frame for %s", frame.GetStream()))
	stream, ok := gs.channelMapping[frame.GetStream()]
	if !ok {
		return &VideoStreamResponse{
			Status: 500,
		}, errors.New("stream not found")
	}
	stream.mu.Lock()
	defer stream.mu.Unlock()
	stream.q <- frame.Shard
	return &VideoStreamResponse{
		Status: 200,
	}, nil
}
