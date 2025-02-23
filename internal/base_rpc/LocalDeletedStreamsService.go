package base_rpc

import "slices"

type LocalDeletedStreamsService struct {
	Deleted []string
}

func NewLocalRemovedStreamsService() DeletedStreamsService {
	return &LocalDeletedStreamsService{
		Deleted: make([]string, 0),
	}
}

func (localRM *LocalDeletedStreamsService) RemoveFromList(streamName string) {
	idx := slices.Index(localRM.Deleted, streamName)
	localRM.Deleted = append(localRM.Deleted[:idx], localRM.Deleted[idx+1:]...)
}

func (localRM *LocalDeletedStreamsService) IsDeleted(streamName string) bool {
	return slices.Contains(localRM.Deleted, streamName)
}

func (localRM *LocalDeletedStreamsService) AddDeletedStream(streamName string) {
	localRM.Deleted = append(localRM.Deleted, streamName)
}
