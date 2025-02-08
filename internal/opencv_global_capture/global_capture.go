package opencv_global_capture

import (
	"github.com/segmentio/kafka-go"
	"go_video_streamer/internal/kafka_consumer"
	"go_video_streamer/internal/rabbitmq_consumer"
	"gocv.io/x/gocv"
)

func NewVideoCapture(streamConfig interface{}) (VideoCapture, error) {
	switch streamConfig.(type) {
	case *rabbitmq_consumer.Config:
		capture, err := rabbitmq_consumer.NewRabbitMQCapture(
			streamConfig.(*rabbitmq_consumer.Config),
		)
		return capture, err
	case *kafka.ReaderConfig:
		capture := kafka_consumer.NewKafkaCapture(streamConfig.(*kafka.ReaderConfig))
		return capture, nil
	}
	capture, err := gocv.OpenVideoCapture(streamConfig)
	return capture, err
}
