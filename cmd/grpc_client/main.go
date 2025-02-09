package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"go_video_streamer/internal/InputStreamShard"
	"go_video_streamer/internal/grpc_consumer"
	"go_video_streamer/internal/video_streaming"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
)

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	serverAddr := os.Getenv("GRPC_ADDR")
	conn, err := grpc.NewClient(serverAddr, opts...)
	if err != nil {
		panic(err)
	}
	defer func(conn *grpc.ClientConn) {
		_ = conn.Close()
	}(conn)

	client := grpc_consumer.NewVideoRCVClient(conn)
	capParams := video_streaming.CaptureParams{
		FPS:    15,
		Width:  960,
		Height: 540,
	}
	captureContext := video_streaming.NewCaptureContext(0, capParams)
	stream, err := client.AddStream(
		context.Background(),
		&grpc_consumer.NewStream{
			Fps:        float32(capParams.FPS),
			Width:      uint32(capParams.Width),
			Height:     uint32(capParams.Height),
			StreamType: 0,
			Name:       "0",
		},
	)
	if err != nil {
		panic(err)
	}
	if stream.Status != 200 {
		log.Panicf("stream.Status: %d", stream.Status)
	}
	for {
		var streamShard InputStreamShard.StreamShard
		img, err := captureContext.Streamer.GetFrame()
		if err != nil {
			log.Fatalln(err)
		}
		streamShard.Gzipped = true
		streamShard.Width = uint32(img.Cols())
		streamShard.Height = uint32(img.Rows())
		streamShard.Fps = float32(captureContext.Streamer.GetVideoFPS())

		var b bytes.Buffer
		zipper := gzip.NewWriter(&b)

		_, err = zipper.Write(img.ToBytes())
		if err != nil {
			panic(err)
		}
		err = zipper.Close()
		if err != nil {
			panic(err)
		}

		streamShard.ImageData = b.Bytes()
		streamShard.MatType = uint32(img.Type())

		newFrame := grpc_consumer.NamedFrame{
			Stream: "0",
			Shard:  &streamShard,
		}
		streamFrame, err := client.StreamFrames(context.Background(), &newFrame)
		if err != nil {
			log.Fatalln(err.Error())
		}
		if streamFrame.Status != 200 {
			log.Printf("streamFrame.Status: %d", streamFrame.Status)
		}
	}
}
