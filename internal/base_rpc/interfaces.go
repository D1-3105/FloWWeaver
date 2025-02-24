package base_rpc

type StreamSchedulerService interface {
	ScheduleDeletion(streamName string, cause *DeletionCause) error
	PerformCheckAndDelete(streamName string) error
}

type DeletedStreamsService interface {
	RemoveFromList(streamName string)
	IsDeleted(streamName string) bool
	AddDeletedStream(streamName string)
}

type ChannelMappingService interface {
	GetStream(s string) (*StreamQ, bool)
	SetStream(name string, stream *StreamQ)
	DeleteStream(name string)
}

type StreamManager interface {
	AddChannelMapping(
		name string, ignoreFailover bool, streamParams *NewStream,
	) (*StreamQ, error)

	RemoveChannel(name string)
	ScheduleDeletion(name string, dc *DeletionCause) error

	GetChannelMapping() ChannelMappingService
	GetManuallyRemoved() DeletedStreamsService
}
