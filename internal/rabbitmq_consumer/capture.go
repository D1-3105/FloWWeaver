package rabbitmq_consumer

import (
	"errors"
	InputStreamShard2 "go_video_streamer/internal/InputStreamShard"
	"gocv.io/x/gocv"
)

type RabbitMQCapture struct {
	props            map[gocv.VideoCaptureProperties]float64
	consumingChannel chan *InputStreamShard2.StreamShard
	consumer         *RMQVideoConsumer
	closed           bool
	closedChan       chan bool
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

func (rc *RabbitMQCapture) IsOpened() bool {
	return !rc.closed
}

func (rc *RabbitMQCapture) Read(mat *gocv.Mat) bool {
	if rc.closed {
		return false
	}
	select {
	case <-rc.closedChan:
		return false
	case shard := <-rc.consumingChannel:
		return InputStreamShard2.ToMatrix(shard, mat)
	}
}

func (rc *RabbitMQCapture) Close() error {
	if rc.closed {
		return errors.New("rmq_consumer already closed")
	}
	rc.closed = true
	rc.closedChan <- true
	rc.consumer.Stop()
	close(rc.consumingChannel)
	return nil
}
