package main

import (
	"bytes"
	"compress/gzip"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	"go_video_streamer/internal/InputStreamShard"
	"go_video_streamer/internal/video_streaming"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"time"
)

func main() {
	opts := stream.NewEnvironmentOptions().SetUri(os.Getenv("RABBIT_STREAM_URI"))
	environment, err := stream.NewEnvironment(opts)
	if err != nil {
		panic(err)
	}
	streamName := os.Getenv("RABBITMQ_QUEUE")
	streamOpts := stream.NewStreamOptions().
		SetMaxAge(time.Minute * 5).
		SetMaxSegmentSizeBytes(stream.ByteCapacity{}.MB(10))
	err = environment.DeclareStream(streamName, streamOpts)
	if err != nil {
		log.Fatal(err)
	}
	prodOpts := stream.NewProducerOptions().SetBatchSize(1).SetProducerName("rabbitmq-stream-producer")
	producer, err := environment.NewProducer(streamName, prodOpts)
	if err != nil {
		panic(err)
	}
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
		log.Printf("Sending message with size %d to rabbitmq", len(msg))
		rmqMessage := amqp.NewMessage(msg)

		go func() {
			err := producer.Send(rmqMessage)
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
