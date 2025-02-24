package kafka_rpc

const (
	KafkaManageUnidentifiedMessage = 0
	KafkaManageNewStream           = 1
	KafkaManageScheduleDeletion    = 2
)

const (
	KafkaManageMessageTypeHeader = "X-FloWWeaver-Message-Type"
)
