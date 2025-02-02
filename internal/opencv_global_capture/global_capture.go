package opencv_global_capture

import (
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
	}
	capture, err := gocv.OpenVideoCapture(streamConfig)
	return capture, err
}
