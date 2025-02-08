package rabbitmq_consumer

import (
	"bytes"
	"compress/gzip"
	"fmt"
	InputStreamShard2 "go_video_streamer/internal/InputStreamShard"
	"gocv.io/x/gocv"
	"io"
	"log/slog"
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
	matData := <-rc.consumingChannel
	var newMat gocv.Mat
	var err error

	if !matData.Gzipped {
		newMat, err = gocv.NewMatFromBytes(
			int(matData.Height),
			int(matData.Width),
			gocv.MatType(matData.MatType),
			matData.ImageData,
		)
	} else {
		reader, err := gzip.NewReader(bytes.NewReader(matData.ImageData))
		if err != nil {
			slog.Error(fmt.Sprintf("Error creating gzip reader: %v", err))
			return false
		}
		defer func(reader *gzip.Reader) {
			_ = reader.Close()
		}(reader)

		decompressedData, err := io.ReadAll(reader)
		if err != nil {
			slog.Error(fmt.Sprintf("Error decompressing image data: %v", err))
			return false
		}

		newMat, err = gocv.NewMatFromBytes(
			int(matData.Height),
			int(matData.Width),
			gocv.MatType(matData.MatType),
			decompressedData,
		)
	}

	if err != nil {
		slog.Error(fmt.Sprintf("Error creating Mat from bytes: %v", err))
		return false
	}

	*mat = newMat
	slog.Debug("Mat read")
	return true
}

func (rc *RabbitMQCapture) Close() error {
	rc.consumer.Stop()
	close(rc.consumingChannel)
	return nil
}
