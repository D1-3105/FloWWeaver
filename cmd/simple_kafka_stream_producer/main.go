package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/segmentio/kafka-go"
	"go_video_streamer/internal/InputStreamShard"
	"go_video_streamer/internal/video_streaming"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"strings"
)

func main() {
	producer := kafka.Writer{
		Addr:     kafka.TCP(strings.Split(os.Getenv("KAFKA_BROKERS"), ",")...),
		Topic:    os.Getenv("KAFKA_TOPIC"),
		Balancer: &kafka.LeastBytes{},
	}
	kafkaK := os.Getenv("KAFKA_GROUP_ID")
	captureContext := video_streaming.NewCaptureContext(0, video_streaming.CaptureParams{
		FPS:    15,
		Width:  960,
		Height: 540,
	})

	go video_streaming.LaunchStreamDaemon(captureContext)
	for {
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
		kafkaMessage := kafka.Message{Key: []byte(kafkaK), Value: msg}
		go func() {
			err := producer.WriteMessages(context.Background(), kafkaMessage)
			if err != nil {
				log.Fatal(err)
			}
			errClose := img.Close()
			if errClose != nil {
				log.Fatal(errClose)
			}
		}()
	}
}
