package msgbus

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/log"
)

const msgbusTopicKey = "msgbus_topic"

// 通道消息
type channelMsg struct {
	topic string
	msg   interface{}
}

type Context struct {
	core.ILogger
	topic string
	msg   interface{}
}

func newContext(msg *channelMsg) core.IMsgbusContext {
	return &Context{
		ILogger: log.Log.NewSessionLogger(zap.String(msgbusTopicKey, msg.topic)),
		topic:   msg.topic,
		msg:     msg.msg,
	}
}
func (c *Context) Topic() string {
	return c.topic
}
func (c *Context) Msg() interface{} {
	return c.msg
}
