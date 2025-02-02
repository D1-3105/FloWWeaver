package opencv_global_capture

import "gocv.io/x/gocv"

type VideoCapture interface {
	Read(img *gocv.Mat) bool
	Set(properties gocv.VideoCaptureProperties, val float64)
	Get(properties gocv.VideoCaptureProperties) float64
	Close() error
}
