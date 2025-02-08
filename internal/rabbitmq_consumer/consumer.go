package rabbitmq_consumer

import (
	"context"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	InputStreamShard2 "go_video_streamer/internal/InputStreamShard"
	"google.golang.org/protobuf/proto"
	"log"
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
		reportChan: make(chan *InputStreamShard2.StreamShard),
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
			err := proto.Unmarshal(message.GetData(), &streamShard)
			if err != nil {
				log.Printf("Error unmarshalling input stream shard data: %v", err)
				return
			}

			log.Println("Message unmarshalled successfully")
			consumer.reportChan <- &streamShard
			err = consumerContext.Consumer.StoreOffset()
			if err != nil {
				panic(err)
			}
		}

		rawConsumer, err := consumer.env.NewConsumer(
			consumer.config.Queue, handler, consumer.config.ConsumerArgs,
		)
		if err != nil {
			log.Printf("Error creating consumer: %v", err)
			consumer.stopChan <- true
			return
		}

		defer func() {
			_ = rawConsumer.Close()
			log.Println("Consumer closed")
		}()
		<-consumerContext.Done()
		log.Println("Consumer goroutine exiting")
	}()

	<-consumer.stopChan
	log.Println("Stopping RMQ Video Consumer")
	cancel()
}

func (consumer *RMQVideoConsumer) Stop() {
	consumer.stopChan <- true
}
