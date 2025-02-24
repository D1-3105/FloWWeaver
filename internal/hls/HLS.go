package hls

import (
	"context"
	"fmt"
	"gocv.io/x/gocv"
	"log/slog"
)

type HLS interface {
	HandleFrame(context.Context, gocv.Mat)
	Dump()
}

type Config struct {
	dir              string
	m3u8Name         string
	frameNumPerShard int
	fps              float64
}

type RepoManager interface {
	WritePatch(string, float64) error
	AddBatch(context.Context, []gocv.Mat) error
	GetConfig() *Config
}

func NewHLSConfig(dir, m3u8Name string, shardDur float64, fps float64) *Config {
	slog.Info(fmt.Sprintf(`New HLSConfig: dir=%s, m3u8=%s, shard=%fs, FPS=%f`, dir, m3u8Name, shardDur, fps))
	return &Config{
		dir:              dir,
		m3u8Name:         m3u8Name,
		fps:              fps,
		frameNumPerShard: int(shardDur * fps),
	}
}
