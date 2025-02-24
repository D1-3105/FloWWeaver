package grpc_consumer

import (
	"context"
	"errors"
	"fmt"
	"go_video_streamer/internal/InputStreamShard"
	baserpc "go_video_streamer/internal/base_rpc"
	"go_video_streamer/internal/opencv_global_capture"
	"log/slog"
)

type GRPCVideoRCVServer struct {
	VideoRCVServer
	baserpc.StreamManager
}

type GRPCCaptureCreator struct {
	Server *GRPCVideoRCVServer
}

func (cc *GRPCCaptureCreator) NewCapture(ch chan *InputStreamShard.StreamShard, _ *baserpc.NewStream) (
	opencv_global_capture.VideoCapture, error) {
	return NewGRPCCapture(
		&ConsumerConfig{grpcServer: cc.Server},
		ch,
	)
}

func NewGRPCVideoRCVServer() *GRPCVideoRCVServer {
	server := &GRPCVideoRCVServer{}
	server.StreamManager = baserpc.NewLocalMemStreamerConsumingService(
		&GRPCCaptureCreator{server},
	)
	return server
}

func (gs *GRPCVideoRCVServer) RMStream(
	_ context.Context, stream *baserpc.RemoveStream,
) (*EditStreamOperationResponse, error) {

	slog.Info(fmt.Sprintf("RMStream initialized via gRPC: name=%s", stream.Name))
	gs.RemoveChannel(stream.Name)
	return &EditStreamOperationResponse{Status: 200, Message: "OK"}, nil
}

func (gs *GRPCVideoRCVServer) AddStream(
	_ context.Context, stream *baserpc.NewStream,
) (*EditStreamOperationResponse, error) {

	slog.Info(
		fmt.Sprintf(
			"New stream initialized via gRPC: name=%s, fps=%f, w=%d, h=%d",
			stream.Name, stream.Fps, stream.Width, stream.Height,
		),
	)

	_, err := gs.AddChannelMapping(stream.Name, false, stream)
	if err == nil {
		return &EditStreamOperationResponse{Status: 200, Message: "OK"}, nil
	} else {
		return &EditStreamOperationResponse{Status: 500, Message: err.Error()}, nil
	}
}

func (gs *GRPCVideoRCVServer) StreamFrames(_ context.Context, frame *NamedFrame) (*VideoStreamResponse, error) {
	slog.Info(fmt.Sprintf("New frame for %s", frame.GetStream()))
	stream, ok := gs.GetChannelMapping().GetStream(frame.GetStream())
	if !ok {
		return &VideoStreamResponse{
			Status: 500,
		}, errors.New("stream not found")
	}
	stream.Mu.Lock()
	defer stream.Mu.Unlock()
	stream.Q <- frame.Shard
	return &VideoStreamResponse{
		Status: 200,
	}, nil
}
