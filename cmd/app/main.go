package main

import (
	"go_video_streamer/internal/video_streaming"
)

func main() {
	ctx := video_streaming.NewCaptureContext("0")
	go video_streaming.LaunchStreamDaemon(ctx)
	video_streaming.ListenStreamToHLS(ctx, "webcam")
}
