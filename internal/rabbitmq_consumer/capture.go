package rabbitmq_consumer

import (
	InputStreamShard2 "go_video_streamer/internal/InputStreamShard"
	"gocv.io/x/gocv"
)

type RabbitMQCapture struct {
	props            map[gocv.VideoCaptureProperties]float64
	consumingChannel chan *InputStreamShard2.StreamShard
	consumer         *RMQVideoConsumer
}

func NewRabbitMQCapture(config *Config) (*RabbitMQCapture, error) {
	consumer := NewRMQVideoConsumer(config)
	go consumer.Start()
	return &RabbitMQCapture{
		props:            make(map[gocv.VideoCaptureProperties]float64),
		consumingChannel: consumer.reportChan,
		consumer:         consumer,
	}, nil
}

func (rc *RabbitMQCapture) Set(vcp gocv.VideoCaptureProperties, value float64) {
	rc.props[vcp] = value
}

func (rc *RabbitMQCapture) Get(vcp gocv.VideoCaptureProperties) float64 {
	value, ok := rc.props[vcp]
	if !ok {
		return 0
	}
	return value
}

func (rc *RabbitMQCapture) Read(mat *gocv.Mat) bool {
	return InputStreamShard2.ToMatrix(&rc.consumer.reportChan, mat)
}

func (rc *RabbitMQCapture) Close() error {
	rc.consumer.Stop()
	close(rc.consumingChannel)
	return nil
}
