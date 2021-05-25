package msgbus

import (
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

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
		ILogger: logger.Log.NewMirrorLogger(msg.topic),
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
