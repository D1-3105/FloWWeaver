package video_streaming

import (
	"context"
	"errors"
	"go_video_streamer/internal/opencv_global_capture"
	"gocv.io/x/gocv"
	"log"
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
			err := (*ctx.Streamer.capture).Close()
			if err != nil {
				log.Printf("Error closing capture Streamer: %v", err)
			}
			return
		}
	}
}

func NewCaptureContext(streamId interface{}, settings CaptureParams) *CaptureContext {
	capture, err := opencv_global_capture.NewVideoCapture(streamId)
	capture.Set(gocv.VideoCaptureFPS, settings.FPS)
	capture.Set(gocv.VideoCaptureFrameWidth, float64(settings.Width))
	capture.Set(gocv.VideoCaptureFrameHeight, float64(settings.Height))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	ctx := CaptureContext{
		Streamer:            NewCaptureStreamer(&capture),
		Context:             context.Background(),
		streamerInitialized: make(chan bool),
	}
	ctx.Value(streamId)
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
		log.Fatal(contextError.Error())
		return contextError
	default:
		return contextError
	}
}

func (ctx *CaptureContext) GetStreamer() *CaptureStreamer {
	return &ctx.Streamer
}
