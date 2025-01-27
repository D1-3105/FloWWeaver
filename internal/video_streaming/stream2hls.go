package video_streaming

import "log"

func ListenStreamToHLS(ctx *CaptureContext, streamName string) {
	streamer := ctx.GetStreamer()

	hlsHandler := NewBaseHLSHandler(NewHLSDirManager(
		NewHLSConfig(
			"stream_repo"+"/"+streamName,
			"index.m3u8",
			5,
			ctx.streamer.GetVideoFPS(),
		)),
	)

	for {
		frame, err := streamer.GetFrame()
		if err != nil || frame.Empty() {
			continue
		}
		log.Println("New frame")

		go hlsHandler.HandleFrame(ctx.Context, frame)
	}
}
