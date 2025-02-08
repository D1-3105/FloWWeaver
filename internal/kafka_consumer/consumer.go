package kafka_consumer

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"go_video_streamer/internal/InputStreamShard"
	"google.golang.org/protobuf/proto"
	"log/slog"
)

type KafkaVideoConsumer struct {
	kafkaReader *kafka.Reader
	stopChan    chan bool
	reportChan  chan *InputStreamShard.StreamShard
}

func NewKafkaVideoConsumer(kafkaConfig *kafka.ReaderConfig) *KafkaVideoConsumer {
	return &KafkaVideoConsumer{
		kafkaReader: kafka.NewReader(*kafkaConfig),
		stopChan:    make(chan bool),
		reportChan:  make(chan *InputStreamShard.StreamShard, 50),
	}
}

func (cons *KafkaVideoConsumer) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		for {
			m, err := cons.kafkaReader.ReadMessage(ctx)
			if err != nil {
				slog.Error(fmt.Sprintf("Error reading message: %v", err))
				continue
			}
			var shardInput InputStreamShard.StreamShard
			err = proto.Unmarshal(m.Value, &shardInput)
			if err != nil {
				slog.Error(fmt.Sprintf("Error unmarshalling message: %v", err))
				continue
			}
			cons.reportChan <- &shardInput
		}
	}(ctx)
	select {
	case <-cons.stopChan:
		err := cons.kafkaReader.Close()
		if err != nil {
			slog.Error(fmt.Sprintf("Error closing kafka reader: %v", err))
		}
		cancel()
		return
	}
}

func (cons *KafkaVideoConsumer) Stop() {
	cons.stopChan <- true
}
