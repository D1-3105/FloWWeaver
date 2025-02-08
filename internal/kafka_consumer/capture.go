package kafka_consumer

import (
	"github.com/segmentio/kafka-go"
	InputStreamShard2 "go_video_streamer/internal/InputStreamShard"
	"gocv.io/x/gocv"
)

type KafkaCapture struct {
	kafkaConfig   *kafka.ReaderConfig
	kafkaConsumer *KafkaVideoConsumer
	props         map[gocv.VideoCaptureProperties]float64
}

func NewKafkaCapture(kafkaConfig *kafka.ReaderConfig) *KafkaCapture {
	kafkaConsumer := NewKafkaVideoConsumer(kafkaConfig)
	go kafkaConsumer.Start()
	return &KafkaCapture{
		kafkaConfig:   kafkaConfig,
		kafkaConsumer: kafkaConsumer,
		props:         make(map[gocv.VideoCaptureProperties]float64),
	}
}

func (k *KafkaCapture) Read(mat *gocv.Mat) bool {
	return InputStreamShard2.ToMatrix(&k.kafkaConsumer.reportChan, mat)
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
	k.kafkaConsumer.Stop()
	close(k.kafkaConsumer.reportChan)
	return nil
}
