package kafka_rpc

import (
	"fmt"
	"github.com/segmentio/kafka-go"
	"go_video_streamer/internal/base_rpc"
	"log/slog"
	"net"
	"strconv"
	"sync"
)

type KafkaChannelMappingService struct {
	base_rpc.LocalChannelMappingService
	leaderConn *kafka.Conn
}

func NewKafkaChannelMappingService(broker string) *KafkaChannelMappingService {
	localChannelMappingService := base_rpc.LocalChannelMappingService{
		ChannelMapping: make(map[string]*base_rpc.StreamQ),
		Mu:             &sync.Mutex{},
	}
	anyConn, err := kafka.Dial("tcp", broker)
	defer func(anyConn *kafka.Conn) {
		_ = anyConn.Close()
	}(anyConn)
	if err != nil {
		panic(err)
	}
	controller, err := anyConn.Controller()
	if err != nil {
		panic(err)
	}
	var connLeader *kafka.Conn
	connLeader, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		panic(err)
	}
	return &KafkaChannelMappingService{
		LocalChannelMappingService: localChannelMappingService,
		leaderConn:                 connLeader,
	}
}

func (kcms *KafkaChannelMappingService) DeleteStream(streamName string) {
	kcms.LocalChannelMappingService.DeleteStream(streamName)
	kcms.Mu.Lock()
	defer kcms.Mu.Unlock()
	err := kcms.leaderConn.DeleteTopics(
		streamName,
	)
	if err != nil {
		slog.Error(fmt.Sprintf("unable to delete stream %s: %v", streamName, err))
	}
}
