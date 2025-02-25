package video_streaming

import (
	"context"
	"fmt"
	"go_video_streamer/internal/hls"
	"gocv.io/x/gocv"
	"log/slog"
)

func ListenStreamToHLS(ctx *CaptureContext, streamName string) {
	streamer := ctx.GetStreamer()
	_, cancel := context.WithCancel(ctx)

	hlsHandler := hls.NewBaseHLSHandler(hls.NewHLSDirManager(
		hls.NewHLSConfig(
			"stream_repo"+"/"+streamName,
			"index.m3u8",
			5,
			ctx.Streamer.GetVideoFPS(),
		)),
	)
	defer func() {
		slog.Info("hls handler stopped - (%s), initializing Dump()", streamName)
		hlsHandler.Dump()
	}()
	defer cancel()

	isStreamInput := int(streamer.capture.Get(gocv.VideoCaptureFrameCount)) <= 0
	for {
		if !(streamer.capture.IsOpened()) {
			slog.Info("Stream %s is not opened", streamName)
			break
		}

		frame, err := streamer.GetFrame()
		if err != nil || frame.Empty() {
			if err != nil {
				slog.Error(fmt.Sprintf("Get frame error: %s", err.Error()))

			} else {
				slog.Info(fmt.Sprintf("Get frame error - captured empty frame: %s", streamName))
			}
			if !isStreamInput {
				break
			}
			continue
		}
		hlsHandler.HandleFrame(ctx.Context, frame)
	}
}

func ListenStreamToHLSWithCallback(ctx *CaptureContext, streamName string, cb func(string) error) {
	defer func() {
		slog.Debug(fmt.Sprintf("HLS Stream (%s) Listener stopped, initializing cb", streamName))
		err := cb(streamName)
		if err != nil {
			slog.Error("ListenStreamToHLSWithCallback: Callback error ", err)
		}
	}()
	ListenStreamToHLS(ctx, streamName)
}
