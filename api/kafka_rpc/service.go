package kafka_rpc

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"go_video_streamer/internal/InputStreamShard"
	baserpc "go_video_streamer/internal/base_rpc"
	"go_video_streamer/internal/kafka_consumer"
	"go_video_streamer/internal/opencv_global_capture"
	"google.golang.org/protobuf/proto"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

func DetectKafkaManagerMessageType(m *kafka.Message) int {
	for _, header := range m.Headers {
		if header.Key == KafkaManageMessageTypeHeader {
			val, err := strconv.Atoi(string(header.Value))
			if err != nil {
				slog.Error("Unable to interpret as integer: %s", err.Error())
			}
			return val
		}
	}
	return KafkaManageUnidentifiedMessage
}

type KafkaCaptureCreator struct {
	Server *KafkaVideoRCV
}

func (kcc *KafkaCaptureCreator) NewCapture(ch chan *InputStreamShard.StreamShard, streamParams *baserpc.NewStream) (
	opencv_global_capture.VideoCapture, error) {
	kapture := kafka_consumer.NewKafkaCapture(
		&kafka.ReaderConfig{
			Topic:   streamParams.Name,
			GroupID: fmt.Sprintf("%f", streamParams.Fps),
			Brokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		},
	)
	kapture.KafkaConsumer.ReportChan = ch
	return kapture, nil
}

type KafkaVideoRCV struct {
	baserpc.StreamManager
	stopChan          chan bool
	kafkaReaderConfig *kafka.ReaderConfig
	KafkaReader       *kafka.Reader
}

func NewKafkaVideoRCV(kafkaReaderConfig *kafka.ReaderConfig) *KafkaVideoRCV {
	kafkaReader := kafka.NewReader(*kafkaReaderConfig)
	rcvServer := KafkaVideoRCV{
		stopChan:          make(chan bool),
		kafkaReaderConfig: kafkaReaderConfig,
		KafkaReader:       kafkaReader,
	}
	manager := baserpc.NewLocalMemStreamerConsumingServiceCustomizable(
		&KafkaCaptureCreator{Server: &rcvServer},
		NewKafkaChannelMappingService(kafkaReaderConfig.Brokers[0]),
	)
	rcvServer.StreamManager = manager
	return &rcvServer
}

func (kvr *KafkaVideoRCV) Start(ctx context.Context) {
	go func() {
		slog.Info(fmt.Sprintf("Started KafkaVideoRCV: %#v\n", kvr.kafkaReaderConfig))
		for {
			m, err := kvr.KafkaReader.ReadMessage(ctx)
			if err != nil {
				slog.Error(fmt.Sprintf("Unable to read message from kafka: %s", err.Error()))
				continue
			}
			switch DetectKafkaManagerMessageType(&m) {
			case KafkaManageNewStream:
				var newStream baserpc.NewStream

				if err := proto.Unmarshal(m.Value, &newStream); err != nil {
					slog.Error(fmt.Sprintf("Unable to unmarshal new stream: %s", err.Error()))
					continue
				}
				slog.Info(fmt.Sprintf("New KafkaVideoRCV.AddChannelMapping: %#v", &newStream))

				_, err = kvr.AddChannelMapping(newStream.Name, true, &newStream)
				if err != nil {
					slog.Error(fmt.Sprintf("Unable to add new stream: %s", err.Error()))
					continue
				}
				break
			case KafkaManageScheduleDeletion:
				var delayedRemoveStream baserpc.DelayedRemoveStream
				if err := proto.Unmarshal(m.Value, &delayedRemoveStream); err != nil {
					slog.Error(fmt.Sprintf("Unable to unmarshal delayed remove stream: %s", err.Error()))
					continue
				}
				slog.Info(fmt.Sprintf("New KafkaVideoRCV.Schedule: %#v", &delayedRemoveStream))
				err = kvr.ScheduleDeletion(delayedRemoveStream.Name, delayedRemoveStream.GetCause())
				if err != nil {
					slog.Error(fmt.Sprintf("Unable to schedule deletion stream: %s", err.Error()))
					continue
				}
				break
			default:
				slog.Error(fmt.Sprintf("Unsupported message type: %s", m.Value))
				continue
			}
		}
	}()
	select {
	case <-ctx.Done():
		return
	case <-kvr.stopChan:
		return
	}
}

func (kvr *KafkaVideoRCV) Stop() {
	kvr.stopChan <- true
}
