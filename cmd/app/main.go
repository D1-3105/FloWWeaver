package main

import (
	"errors"
	"fmt"
	"go_video_streamer/internal/video_streaming"
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
		http.Handle("/hls/", http.StripPrefix("/hls/", http.FileServer(http.Dir("./stream_repo/webcam/"))))
		err := http.ListenAndServe(":8080", nil)
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server closed\n")
		} else if err != nil {
			fmt.Printf("error starting server: %s\n", err)
			os.Exit(1)
		}
	}
}
