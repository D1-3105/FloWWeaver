package main

import (
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"go_video_streamer/internal/video_streaming"
	"log/slog"
	"net/http"
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
