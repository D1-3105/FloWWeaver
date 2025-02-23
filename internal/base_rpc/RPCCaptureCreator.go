package base_rpc

import (
	"go_video_streamer/internal/InputStreamShard"
	"go_video_streamer/internal/opencv_global_capture"
)

type CaptureCreator interface {
	NewCapture(ch chan *InputStreamShard.StreamShard) (opencv_global_capture.VideoCapture, error)
}
