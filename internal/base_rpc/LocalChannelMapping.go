package base_rpc

import (
	"go_video_streamer/internal/InputStreamShard"
	"go_video_streamer/internal/video_streaming"
	"sync"
)

type StreamQ struct {
	Q              chan *InputStreamShard.StreamShard
	Mu             sync.Locker
	CaptureContext *video_streaming.CaptureContext
}

type LocalChannelMappingService struct {
	ChannelMapping map[string]*StreamQ
}

func (c *LocalChannelMappingService) GetStream(s string) (*StreamQ, bool) {
	stream, ok := c.ChannelMapping[s]
	return stream, ok
}

func (c *LocalChannelMappingService) SetStream(name string, stream *StreamQ) {
	c.ChannelMapping[name] = stream
}

func (c *LocalChannelMappingService) DeleteStream(name string) {
	delete(c.ChannelMapping, name)
}

func NewChannelMappingMap() ChannelMappingService {
	return &LocalChannelMappingService{
		ChannelMapping: make(map[string]*StreamQ),
	}
}
