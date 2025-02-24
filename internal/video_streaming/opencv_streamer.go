package video_streaming

import (
	"errors"
	"go_video_streamer/internal/opencv_global_capture"
	"gocv.io/x/gocv"
	"log/slog"
	"sync"
	"sync/atomic"
)

type CaptureError struct {
	Err error
}

func (e CaptureError) Error() string {
	return e.Err.Error()
}

type CaptureStreamer struct {
	capture      opencv_global_capture.VideoCapture
	captureMutex sync.Locker
	Stats        *StreamerStats
}

func NewCaptureStreamer(capture opencv_global_capture.VideoCapture) CaptureStreamer {
	streamer := CaptureStreamer{
		capture:      capture,
		captureMutex: &sync.Mutex{},
		Stats:        NewStreamerStats(),
	}

	streamer.Stats.HandledFrames = atomic.Uint64{}
	streamer.Stats.HandledFrames.Store(0)
	streamer.Stats.targetStreamFramesCount = atomic.Uint64{}
	streamer.Stats.isTargetFrames.Store(false)
	return streamer
}

func (streamer *CaptureStreamer) SetTargetStreamFramesCount(cnt uint64) {
	streamer.Stats.targetStreamFramesCount.Store(cnt)
	streamer.Stats.isTargetFrames.Store(true)
}

func (streamer *CaptureStreamer) GetFrame() (gocv.Mat, error) {
	streamer.captureMutex.Lock()
	defer streamer.captureMutex.Unlock()

	img := gocv.NewMat()
	err := !streamer.capture.Read(&img)
	if err {
		return img, CaptureError{Err: errors.New("failed to read frame")}
	}
	streamer.Stats.HandledFrames.Add(1)
	if streamer.Stats.ShouldBeReaped() {
		slog.Debug("streamer should be reaped")
		err := streamer.capture.Close()
		if err != nil {
			slog.Error(err.Error())
		} else {
			slog.Debug("streamer.capture closed!")
		}
	}
	return img, nil
}

func (streamer *CaptureStreamer) GetVideoFPS() float64 {
	return streamer.capture.Get(gocv.VideoCaptureFPS)
}
