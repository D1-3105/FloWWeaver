package main

import (
	"context"
	"github.com/segmentio/kafka-go"
	"go_video_streamer/api/file_server"
	"go_video_streamer/api/kafka_rpc"
	utils_cmd "go_video_streamer/cmd/utils"
	"os"
	"strings"
)

func main() {
	utils_cmd.SetupSlog()
	topic, exists := os.LookupEnv("FLOWWEAVER_GENERAL_TOPIC")
	if !exists {
		topic = "flowweaver-general"
	}
	brokers, exists := os.LookupEnv("KAFKA_BROKERS")
	if !exists {
		brokers = "localhost:9092"
	}
	brokersList := strings.Split(brokers, ",")
	groupId, exists := os.LookupEnv("FLOWWEAVER_GENERAL_GROUP_ID")
	if !exists {
		groupId = "all"
	}
	readerConfig := kafka.ReaderConfig{
		Brokers: brokersList,
		Topic:   topic,
		GroupID: groupId,
	}
	kafkaManager := kafka_rpc.NewKafkaVideoRCV(
		&readerConfig,
	)
	ctx, cancel := context.WithCancel(context.Background())
	go kafkaManager.Start(ctx)
	if os.Getenv("SERVER_OFF") != "true" {
		go file_server.LaunchStreamRepoFileServer(ctx)
	}
	select {
	case <-ctx.Done():
		kafkaManager.Stop()
		cancel()
		return
	}
}
