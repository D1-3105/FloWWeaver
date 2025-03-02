package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"go_video_streamer/api/kafka_rpc"
	"go_video_streamer/internal/InputStreamShard"
	"go_video_streamer/internal/base_rpc"
	"go_video_streamer/internal/video_streaming"
	"google.golang.org/protobuf/proto"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

func main() {
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
	conn, err := kafka.DialLeader(
		context.Background(), "tcp", brokersList[0], topic, 0,
	)
	if err != nil {
		slog.Error(fmt.Sprintf("Error during connection to leader: %s", err))
		panic(err)
	}
	err = conn.CreateTopics(
		kafka.TopicConfig{Topic: topic, NumPartitions: 1, ReplicationFactor: 1},
		kafka.TopicConfig{Topic: "0", NumPartitions: 1, ReplicationFactor: 1},
	)
	if err != nil {
		panic(err)
	}

	producerGeneral := kafka.Writer{
		Addr:     kafka.TCP(brokersList...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer func(producerGeneral *kafka.Writer) {
		_ = producerGeneral.Close()
	}(&producerGeneral)

	producerStream := kafka.Writer{
		Addr:     kafka.TCP(brokersList...),
		Topic:    "0",
		Balancer: &kafka.LeastBytes{},
	}
	defer func(producerGeneral *kafka.Writer) {
		_ = producerGeneral.Close()
	}(&producerGeneral)

	capParams := video_streaming.CaptureParams{
		FPS:    15,
		Width:  960,
		Height: 540,
	}
	captureContext := video_streaming.NewCaptureContext(0, capParams)
	newStream := base_rpc.NewStream{
		Fps:        float32(capParams.FPS),
		Width:      uint32(capParams.Width),
		Height:     uint32(capParams.Height),
		StreamType: 0,
		Name:       "0",
	}
	serialized, err := proto.Marshal(&newStream)
	if err != nil {
		panic(err)
	}
	err = producerGeneral.WriteMessages(
		context.Background(), kafka.Message{
			Key:   []byte(groupId),
			Value: serialized,
			Headers: []kafka.Header{
				{
					Key:   kafka_rpc.KafkaManageMessageTypeHeader,
					Value: []byte(fmt.Sprintf("%d", kafka_rpc.KafkaManageNewStream)),
				},
			},
		},
	)
	if err != nil {
		panic(err)
	}
	deleteStreamDelayed := base_rpc.DelayedRemoveStream{
		Name:  newStream.Name,
		Cause: &base_rpc.DeletionCause{TargetFrames: 141},
	}
	go func() {
		serialized, err = proto.Marshal(&deleteStreamDelayed)
		if err != nil {
			panic(err)
		}
		err := producerGeneral.WriteMessages(
			context.Background(),
			kafka.Message{
				Key: []byte(groupId), Value: serialized,
				Headers: []kafka.Header{
					{
						Key:   kafka_rpc.KafkaManageMessageTypeHeader,
						Value: []byte(fmt.Sprintf("%d", kafka_rpc.KafkaManageScheduleDeletion)),
					},
				},
			},
		)
		if err != nil {
			slog.Error(err.Error())
		}
	}()

	go video_streaming.LaunchStreamDaemon(captureContext)
	sent := atomic.Uint64{}
	sent.Store(0)
	for range deleteStreamDelayed.Cause.TargetFrames {
		var streamShard InputStreamShard.StreamShard
		img, err := captureContext.Streamer.GetFrame()
		if err != nil {
			log.Fatalln(err)
		}
		streamShard.Gzipped = true
		streamShard.Width = uint32(img.Cols())
		streamShard.Height = uint32(img.Rows())
		streamShard.Fps = float32(captureContext.Streamer.GetVideoFPS())

		var b bytes.Buffer
		zipper := gzip.NewWriter(&b)

		_, err = zipper.Write(img.ToBytes())
		if err != nil {
			panic(err)
		}
		err = zipper.Close()
		if err != nil {
			panic(err)
		}

		streamShard.ImageData = b.Bytes()
		streamShard.MatType = uint32(img.Type())

		msg, err := proto.Marshal(&streamShard)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Sending message with size %d to kafka", len(msg))
		kafkaMessage := kafka.Message{Key: []byte(fmt.Sprintf("%f", capParams.FPS)), Value: msg}
		go func() {
			err := producerStream.WriteMessages(context.Background(), kafkaMessage)
			sent.Add(1)
			if err != nil {
				log.Fatal(err)
			}
			errClose := img.Close()
			if errClose != nil {
				log.Fatal(errClose)
			}
		}()
	}
	for {
		now := sent.Load()
		if now == deleteStreamDelayed.Cause.TargetFrames {
			break
		}
		slog.Info(fmt.Sprintf("Sent %d of %d", now, deleteStreamDelayed.Cause.TargetFrames))
		time.Sleep(2)
	}
}
