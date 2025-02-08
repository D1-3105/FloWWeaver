package main

import (
	"errors"
	"go_video_streamer/internal/video_streaming"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	ctx := video_streaming.NewCaptureContext("0", video_streaming.CaptureParams{
		FPS:    10,
		Width:  1280,
		Height: 720,
	})
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
			slog.Info("server closed")
		} else if err != nil {
			slog.Error("error starting server: %s", err)
			os.Exit(1)
		}
	}
}
