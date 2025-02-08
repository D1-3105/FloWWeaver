package hls

import (
	"context"
	"fmt"
	"gocv.io/x/gocv"
	"log"
	"log/slog"
	"sync"
)

type BaseHLSHandler struct {
	config           *Config
	streamDirManager RepoManager
	frameQ           *[]gocv.Mat
	frameCnt         int
	mutex            sync.Mutex
}

func (h *BaseHLSHandler) HandleFrame(ctx context.Context, frame gocv.Mat) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	(*h.frameQ)[h.frameCnt] = frame
	h.frameCnt++
	slog.Debug(fmt.Sprintf("Frame: %d of %d", h.frameCnt, h.config.frameNumPerShard))
	if h.frameCnt == h.config.frameNumPerShard {
		currentBatch := *h.frameQ
		go func() {
			log.Println("New HLS shard enqueued")
			err := h.streamDirManager.AddBatch(ctx, currentBatch)
			if err != nil {
				panic(err)
			}
		}()
		newFrameQ := make([]gocv.Mat, h.config.frameNumPerShard)
		h.frameQ = &newFrameQ
		h.frameCnt = 0
	}
}

func NewBaseHLSHandler(manager RepoManager) HLS {
	data := make([]gocv.Mat, manager.GetConfig().frameNumPerShard)
	return &BaseHLSHandler{
		config:           manager.GetConfig(),
		streamDirManager: manager,
		frameQ:           &data,
	}
}
