package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	"go_video_streamer/api/file_server"
	"go_video_streamer/internal/rabbitmq_consumer"
	"go_video_streamer/internal/video_streaming"
	"os"
	"time"
)

func main() {
	consOptions := stream.NewConsumerOptions().
		SetConsumerName(fmt.Sprintf("stream-consumer-%v", uuid.New().String())).
		SetCRCCheck(false).
		SetOffset(stream.OffsetSpecification{}.First()).
		SetManualCommit()
	streamOptions := stream.NewStreamOptions().
		SetMaxAge(time.Minute * 5).
		SetMaxSegmentSizeBytes(stream.ByteCapacity{}.MB(10))
	ctx := video_streaming.NewCaptureContext(&rabbitmq_consumer.Config{
		ConsumerArgs:   consOptions,
		Queue:          os.Getenv("RABBITMQ_QUEUE"),
		PrefetchCount:  1,
		QueueType:      rabbitmq_consumer.StreamQ,
		QueueExtraArgs: streamOptions,
		URI:            os.Getenv("RABBIT_STREAM_URI"),
	},
		video_streaming.CaptureParams{
			FPS:    15,
			Width:  960,
			Height: 540,
		},
	)
	go video_streaming.LaunchStreamDaemon(ctx)

	if os.Getenv("SERVER_OFF") == "true" {
		video_streaming.ListenStreamToHLS(ctx, "webcam")
	} else {
		go video_streaming.ListenStreamToHLS(ctx, "webcam")
		file_server.LaunchStreamRepoFileServer(ctx)
	}
}
