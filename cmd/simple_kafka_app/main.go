package main

import (
	"github.com/segmentio/kafka-go"
	"go_video_streamer/api/file_server"
	"go_video_streamer/internal/video_streaming"
	"os"
	"strings"
)

func main() {
	ctx := video_streaming.NewCaptureContext(&kafka.ReaderConfig{
		Brokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		Topic:   os.Getenv("KAFKA_TOPIC"),
		GroupID: os.Getenv("KAFKA_GROUP_ID"),
	}, video_streaming.CaptureParams{
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
