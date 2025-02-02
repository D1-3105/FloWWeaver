package rabbitmq_consumer

import (
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	"strings"
)

type QType string

const (
	StreamQ QType = "stream"
	Classic QType = "classic"
	Quorum  QType = "quorum"
)

type Config struct {
	URI            string
	Queue          string
	PrefetchCount  int
	QueueType      QType
	ConsumerArgs   *stream.ConsumerOptions
	QueueExtraArgs *stream.StreamOptions
}

func (c *Config) ToStreamEnv() *stream.Environment {
	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().
			SetUri(c.URI).IsTLS(strings.HasPrefix(c.URI, "rabbitmq-stream+tls://")),
	)
	if err != nil {
		panic(err)
	}
	err = env.DeclareStream(c.Queue, c.QueueExtraArgs)
	if err != nil {
		panic(err)
	}
	return env
}
