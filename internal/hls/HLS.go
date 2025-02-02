package hls

import (
	"context"
	"gocv.io/x/gocv"
	"log"
)

type HLS interface {
	HandleFrame(context.Context, gocv.Mat)
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
	log.Printf(`New HLSConfig: dir=%s, m3u8=%s, shard=%fs, FPS=%f`, dir, m3u8Name, shardDur, fps)
	return &Config{
		dir:              dir,
		m3u8Name:         m3u8Name,
		fps:              fps,
		frameNumPerShard: int(shardDur * fps),
	}
}
