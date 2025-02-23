package base_rpc

import (
	"context"
	"errors"
	"fmt"
	"go_video_streamer/internal/InputStreamShard"
	"go_video_streamer/internal/video_streaming"
	"sync"
)

type StreamerManagerService struct {
	ChannelMapping   ChannelMappingService
	ManuallyRemoved  DeletedStreamsService
	EditStreamsMutex sync.Locker
	CapCreator       CaptureCreator
}

func (scs *StreamerManagerService) GetManuallyRemoved() DeletedStreamsService {
	return scs.ManuallyRemoved
}

func (scs *StreamerManagerService) GetChannelMapping() ChannelMappingService {
	return scs.ChannelMapping
}

func NewLocalMemStreamerConsumingService(creator CaptureCreator) StreamManager {
	return &StreamerManagerService{
		ChannelMapping:   NewChannelMappingMap(),
		ManuallyRemoved:  NewLocalRemovedStreamsService(),
		EditStreamsMutex: &sync.Mutex{},
		CapCreator:       creator,
	}
}

func (scs *StreamerManagerService) AddChannelMapping(
	name string, ignoreFailover bool, streamParams *NewStream,
) (*StreamQ, error) {
	scs.EditStreamsMutex.Lock()
	defer scs.EditStreamsMutex.Unlock()
	if scs.ManuallyRemoved.IsDeleted(name) {
		if ignoreFailover {
			scs.ManuallyRemoved.RemoveFromList(name)
		} else {
			return nil, fmt.Errorf("channel already removed")
		}
	}
	channel := make(chan *InputStreamShard.StreamShard, 50)
	capture, err := scs.CapCreator.NewCapture(channel)
	if err != nil {
		panic(err)
	}
	streamWrapper := StreamQ{
		Q:  channel,
		Mu: &sync.Mutex{},
		CaptureContext: video_streaming.CaptureContextFromCapture(
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
		go video_streaming.ListenStreamToHLS(streamWrapper.CaptureContext, name)
		break

	default:
		panic(errors.New("invalid stream type"))
	}

	scs.ChannelMapping.SetStream(name, &streamWrapper)
	return &streamWrapper, nil
}

func (scs *StreamerManagerService) RemoveChannel(name string) {
	scs.EditStreamsMutex.Lock()
	defer scs.EditStreamsMutex.Unlock()
	stream, ok := scs.ChannelMapping.GetStream(name)
	if !ok {
		return
	}
	_, cancel := context.WithCancel(stream.CaptureContext.Context)
	cancel()
	scs.ChannelMapping.DeleteStream(name)
	if !scs.ManuallyRemoved.IsDeleted(name) {
		scs.ManuallyRemoved.AddDeletedStream(name)
	}
}
