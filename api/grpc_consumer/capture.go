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
}

func (c *GRPCCapture) Read(mat *gocv.Mat) bool {
	return InputStreamShard.ToMatrix(&c.listenChannel, mat)
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
	return nil
}

func NewGRPCCapture(config *ConsumerConfig, listenChannel chan *InputStreamShard.StreamShard) (opencv_global_capture.VideoCapture, error) {
	return &GRPCCapture{
		config:        config,
		listenChannel: listenChannel,
		props:         make(map[gocv.VideoCaptureProperties]float64),
	}, nil
}
