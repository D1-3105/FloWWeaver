package main

import "gocv.io/x/gocv"

func main() {
	capture, err := gocv.OpenVideoCapture("http://localhost:8080/hls/webcam/index.m3u8")
	if err != nil {
		panic(err)
	}
	defer func(capture *gocv.VideoCapture) {
		_ = capture.Close()
	}(capture)

	window := gocv.NewWindow("HLS stream")
	window.SetWindowProperty(gocv.WindowPropertyAutosize, gocv.WindowAutosize)

	for {
		img := gocv.NewMat()
		if ok := capture.Read(&img); !ok || img.Empty() {
			break
		}
		window.IMShow(img)
		if gocv.WaitKey(1) >= 0 {
			break
		}
	}
}
