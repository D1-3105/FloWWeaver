package kafka_consumer

import (
	"context"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"go_video_streamer/internal/InputStreamShard"
	"google.golang.org/protobuf/proto"
	"io"
	"log/slog"
)

type KafkaVideoConsumer struct {
	kafkaReader *kafka.Reader
	stopChan    chan bool
	ReportChan  chan *InputStreamShard.StreamShard
}

func NewKafkaVideoConsumer(kafkaConfig *kafka.ReaderConfig) *KafkaVideoConsumer {
	return &KafkaVideoConsumer{
		kafkaReader: kafka.NewReader(*kafkaConfig),
		stopChan:    make(chan bool),
		ReportChan:  make(chan *InputStreamShard.StreamShard, 1),
	}
}

func (cons *KafkaVideoConsumer) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				m, err := cons.kafkaReader.ReadMessage(ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						slog.Info("Kafka consumer stopped: context canceled")
						return
					}
					if errors.Is(err, io.EOF) {
						slog.Info("Kafka consumer stopped: EOF received")
						return
					}
					slog.Error(fmt.Sprintf("Error reading message: %v", err))
					continue
				}
				var shardInput InputStreamShard.StreamShard
				err = proto.Unmarshal(m.Value, &shardInput)
				if err != nil {
					slog.Error(fmt.Sprintf("Error unmarshalling message: %v", err))
					continue
				}
				cons.ReportChan <- &shardInput
			}
		}
	}(ctx)
	select {
	case <-cons.stopChan:
		cancel()
	case <-ctx.Done():
	}
	if err := cons.kafkaReader.Close(); err != nil {
		slog.Error(fmt.Sprintf("Error closing kafka reader: %v", err))
	}
}

func (cons *KafkaVideoConsumer) Stop() {
	cons.stopChan <- true
}
