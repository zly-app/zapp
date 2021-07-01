package msgbus_consume

import (
	"github.com/zly-app/zapp/core"
)

type ConsumerConfig struct {
	IsGlobal        bool
	Topic           string
	ThreadCount     int
	Handler         core.MsgbusHandler
	ConsumeAttempts uint16 // 消费尝试次数, 默认3, 最大65535
}
