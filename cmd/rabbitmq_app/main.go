package main

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	"go_video_streamer/internal/rabbitmq_consumer"
	"go_video_streamer/internal/video_streaming"
	"log/slog"
	"net/http"
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

		http.HandleFunc("/hls/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == http.MethodOptions {
				return
			}

			http.StripPrefix("/hls/", http.FileServer(http.Dir("./stream_repo/"))).ServeHTTP(w, r)
		})

		err := http.ListenAndServe(":8080", nil)
		if errors.Is(err, http.ErrServerClosed) {
			slog.Info("server closed\n")
		} else if err != nil {
			slog.Error(fmt.Sprintf("error starting server: %s\n", err))
			os.Exit(1)
		}
	}
}
