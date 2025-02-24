package grpc_consumer

import (
	"go_video_streamer/internal/InputStreamShard"
	"go_video_streamer/internal/opencv_global_capture"
	"gocv.io/x/gocv"
)

type GRPCCapture struct {
	config        *ConsumerConfig
	listenChannel chan *InputStreamShard.StreamShard
	props         map[gocv.VideoCaptureProperties]float64
	closed        bool
	closedChan    chan bool
}

func (c *GRPCCapture) Read(mat *gocv.Mat) bool {
	if c.closed {
		return false
	}
	select {
	case <-c.closedChan:
		return false
	case shard := <-c.listenChannel:
		return InputStreamShard.ToMatrix(shard, mat)
	}
}

func (c *GRPCCapture) Set(vcp gocv.VideoCaptureProperties, value float64) {
	c.props[vcp] = value
}

func (c *GRPCCapture) Get(vcp gocv.VideoCaptureProperties) float64 {
	value, ok := c.props[vcp]
	if !ok {
		return 0
	}
	return value
}

func (c *GRPCCapture) Close() error {
	c.config.grpcServer.RemoveChannel(c.config.name)
	c.closed = true
	c.closedChan <- true
	return nil
}

func NewGRPCCapture(config *ConsumerConfig, listenChannel chan *InputStreamShard.StreamShard) (opencv_global_capture.VideoCapture, error) {
	return &GRPCCapture{
		config:        config,
		listenChannel: listenChannel,
		props:         make(map[gocv.VideoCaptureProperties]float64),
		closed:        false,
		closedChan:    make(chan bool),
	}, nil
}

func (c *GRPCCapture) IsOpened() bool {
	return !c.closed
}
