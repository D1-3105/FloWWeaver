package kafka_consumer

import (
	"errors"
	"github.com/segmentio/kafka-go"
	InputStreamShard2 "go_video_streamer/internal/InputStreamShard"
	"gocv.io/x/gocv"
)

type KafkaCapture struct {
	kafkaConfig   *kafka.ReaderConfig
	KafkaConsumer *KafkaVideoConsumer
	props         map[gocv.VideoCaptureProperties]float64
	closed        bool
	closedChan    chan bool
}

func NewKafkaCapture(kafkaConfig *kafka.ReaderConfig) *KafkaCapture {
	kafkaConsumer := NewKafkaVideoConsumer(kafkaConfig)
	go kafkaConsumer.Start()
	return &KafkaCapture{
		kafkaConfig:   kafkaConfig,
		KafkaConsumer: kafkaConsumer,
		props:         make(map[gocv.VideoCaptureProperties]float64),
		closed:        false,
		closedChan:    make(chan bool, 10),
	}
}

func (k *KafkaCapture) Read(mat *gocv.Mat) bool {
	if k.closed {
		return false
	}
	select {
	case <-k.closedChan:
		return false
	case shard := <-k.KafkaConsumer.ReportChan:
		return InputStreamShard2.ToMatrix(shard, mat)
	}
}

func (k *KafkaCapture) Set(vcp gocv.VideoCaptureProperties, value float64) {
	k.props[vcp] = value
}

func (k *KafkaCapture) Get(vcp gocv.VideoCaptureProperties) float64 {
	value, ok := k.props[vcp]
	if !ok {
		return 0
	}
	return value
}

func (k *KafkaCapture) Close() error {
	if k.closed {
		return errors.New("kafka_consumer already closed")
	}
	k.KafkaConsumer.Stop()
	k.closed = true
	k.closedChan <- true
	close(k.KafkaConsumer.ReportChan)
	return nil
}

func (k *KafkaCapture) IsOpened() bool {
	return !k.closed
}
