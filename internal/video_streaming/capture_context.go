package video_streaming

import (
	"context"
	"errors"
	"fmt"
	"go_video_streamer/internal/opencv_global_capture"
	"gocv.io/x/gocv"
	"log/slog"
)

type CaptureContext struct {
	context.Context

	Streamer            CaptureStreamer
	streamerInitialized chan bool
}

type CaptureParams struct {
	FPS    float64
	Width  int
	Height int
}

func LaunchStreamDaemon(ctx *CaptureContext) {
	for {
		select {
		case <-ctx.Done():
			err := ctx.Streamer.capture.Close()
			if err != nil {
				slog.Error(fmt.Sprintf(
					"Error closing capture Streamer inside of LaunchStreamDaemon: %v", err,
				))
			}
			return
		}
	}
}

func NewCaptureContext(streamId interface{}, settings CaptureParams) *CaptureContext {
	capture, err := opencv_global_capture.NewVideoCapture(streamId)
	if err != nil {
		slog.Error(err.Error())
		return nil
	}
	capture.Set(gocv.VideoCaptureFPS, settings.FPS)
	capture.Set(gocv.VideoCaptureFrameWidth, float64(settings.Width))
	capture.Set(gocv.VideoCaptureFrameHeight, float64(settings.Height))
	ctx := CaptureContext{
		Streamer:            NewCaptureStreamer(capture),
		Context:             context.Background(),
		streamerInitialized: make(chan bool),
	}
	ctx.Value(streamId)
	go LaunchStreamDaemon(&ctx)
	return &ctx
}

func CaptureContextFromCapture(capture opencv_global_capture.VideoCapture, settings CaptureParams, name string) *CaptureContext {
	capture.Set(gocv.VideoCaptureFPS, settings.FPS)
	capture.Set(gocv.VideoCaptureFrameWidth, float64(settings.Width))
	capture.Set(gocv.VideoCaptureFrameHeight, float64(settings.Height))
	ctx := CaptureContext{
		Streamer:            NewCaptureStreamer(capture),
		Context:             context.Background(),
		streamerInitialized: make(chan bool),
	}
	ctx.Value(name)
	go LaunchStreamDaemon(&ctx)
	return &ctx
}

func (ctx *CaptureContext) Done() <-chan struct{} {
	return ctx.Context.Done()
}

func (ctx *CaptureContext) Err() error {
	contextError := ctx.Context.Err()
	var captureError *CaptureError
	switch {
	case errors.As(contextError, &captureError):
		slog.Error(contextError.Error())
		return contextError
	default:
		return contextError
	}
}

func (ctx *CaptureContext) GetStreamer() *CaptureStreamer {
	return &ctx.Streamer
}
