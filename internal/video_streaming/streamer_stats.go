package video_streaming

import (
	"fmt"
	"log/slog"
	"sync/atomic"
)

type StreamerStats struct {
	HandledFrames           atomic.Uint64
	targetStreamFramesCount atomic.Uint64
	isTargetFrames          atomic.Bool
}

func NewStreamerStats() *StreamerStats {
	return &StreamerStats{
		HandledFrames:           atomic.Uint64{},
		targetStreamFramesCount: atomic.Uint64{},
		isTargetFrames:          atomic.Bool{},
	}
}

func (s *StreamerStats) ShouldBeReaped() bool {
	isTargetFrames := s.isTargetFrames.Load()
	isReached := s.targetStreamFramesCount.Load() <= s.HandledFrames.Load()
	slog.Debug(fmt.Sprintf("isTargetFrames: %t && isReached: %t", isTargetFrames, isReached))
	return isTargetFrames && isReached
}
