package video_streaming

import (
	"go_video_streamer/internal/hls"
	"log"
)

func ListenStreamToHLS(ctx *CaptureContext, streamName string) {
	streamer := ctx.GetStreamer()

	hlsHandler := hls.NewBaseHLSHandler(hls.NewHLSDirManager(
		hls.NewHLSConfig(
			"stream_repo"+"/"+streamName,
			"index.m3u8",
			5,
			ctx.Streamer.GetVideoFPS(),
		)),
	)

	for {
		frame, err := streamer.GetFrame()
		if err != nil || frame.Empty() {
			log.Println("Get frame error", err)
			continue
		}
		go hlsHandler.HandleFrame(ctx.Context, frame)
	}
}
