package video_streaming

import (
	"context"
	"gocv.io/x/gocv"
)

type HLS interface {
	HandleFrame(context.Context, gocv.Mat)
}

type HLSConfig struct {
	dir              string
	m3u8Name         string
	frameNumPerShard int
	fps              float64
}

type HLSRepoManager interface {
	WritePatch(string, float64) error
	AddBatch(context.Context, []gocv.Mat) error
	GetConfig() *HLSConfig
}

func NewHLSConfig(dir, m3u8Name string, shardDur float64, fps float64) *HLSConfig {
	return &HLSConfig{
		dir:              dir,
		m3u8Name:         m3u8Name,
		fps:              fps,
		frameNumPerShard: int(shardDur * fps),
	}
}
