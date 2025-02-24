package rabbitmq_consumer

import (
	"context"
	"fmt"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	InputStreamShard2 "go_video_streamer/internal/InputStreamShard"
	"google.golang.org/protobuf/proto"
	"log/slog"
)

type RMQVideoConsumer struct {
	reportChan chan *InputStreamShard2.StreamShard
	stopChan   chan bool
	config     *Config
	env        *stream.Environment
}

func NewRMQVideoConsumer(config *Config) *RMQVideoConsumer {
	if config.QueueType != StreamQ {
		panic("rabbitmq_consumer can only be used with StreamQ")
	}
	env := config.ToStreamEnv()
	return &RMQVideoConsumer{
		reportChan: make(chan *InputStreamShard2.StreamShard, 50),
		config:     config,
		stopChan:   make(chan bool),
		env:        env,
	}
}

func (consumer *RMQVideoConsumer) Start() {
	consumerContext, cancel := context.WithCancel(context.Background())

	go func() {
		handler := func(consumerContext stream.ConsumerContext, message *amqp.Message) {
			var streamShard InputStreamShard2.StreamShard
			if err := proto.Unmarshal(message.GetData(), &streamShard); err != nil {
				slog.Error(fmt.Sprintf("Error unmarshalling input stream shard data: %v", err))
				return
			}

			slog.Debug("Message unmarshalled successfully")
			consumer.reportChan <- &streamShard

		}

		rawConsumer, err := consumer.env.NewConsumer(
			consumer.config.Queue, handler, consumer.config.ConsumerArgs,
		)
		if err != nil {
			slog.Error(fmt.Sprintf("Error creating consumer: %v", err))
			consumer.stopChan <- true
			return
		}

		defer func() {
			_ = rawConsumer.Close()
			slog.Info("Consumer closed")
		}()
		<-consumerContext.Done()
		slog.Info("Consumer goroutine exiting")
	}()

	<-consumer.stopChan
	slog.Info("Stopping RMQ Video Consumer")
	cancel()
}

func (consumer *RMQVideoConsumer) Stop() {
	consumer.stopChan <- true
}
