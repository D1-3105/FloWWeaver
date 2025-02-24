package base_rpc

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type LocalStreamScheduler struct {
	scheduled            ChannelMappingService
	deleteStreamsDelayed DeletedStreamsService

	mu sync.Locker
}

func NewLocalStreamScheduler(chService ChannelMappingService) *LocalStreamScheduler {
	return &LocalStreamScheduler{
		chService,
		NewLocalRemovedStreamsService(),
		&sync.Mutex{},
	}
}

func (s *LocalStreamScheduler) ScheduleDeletion(streamName string, cause *DeletionCause) error {
	slog.Info(fmt.Sprintf("Scheduling deletion of (%s), %+v", streamName, cause))
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteStreamsDelayed.AddDeletedStream(streamName)
	go func() {
		for {
			stream, ok := s.scheduled.GetStream(streamName)
			if ok {
				stream.CaptureContext.Streamer.SetTargetStreamFramesCount(cause.TargetFrames)
				slog.Info(fmt.Sprintf("SetTargetStreamFramesCount (%s), %+v", streamName, cause))
				break
			} else {
				slog.Debug(fmt.Sprintf("Stream %s not found", streamName))
				time.Sleep(time.Second * 100)
				continue
			}
		}
	}()
	return nil
}

func (s *LocalStreamScheduler) PerformCheckAndDelete(streamName string) error {
	scheduledStreamQ, ok := s.scheduled.GetStream(streamName)
	if !ok {
		slog.Error("Stream '%s' not found during check", streamName)
		return errors.New("stream not found during check")
	}
	isDeletable := scheduledStreamQ.CaptureContext.Streamer.Stats.ShouldBeReaped()
	if isDeletable && s.deleteStreamsDelayed.IsDeleted(streamName) {
		s.scheduled.DeleteStream(streamName)
	}
	return nil
}
