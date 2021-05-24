package msgbus

import (
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

type Context struct {
	core.ILogger
	topic string
	msg   interface{}
}

func newContext(topic string, msg interface{}) core.IMsgbusContext {
	return &Context{
		ILogger: logger.Log.NewMirrorLogger(topic),
		topic:   topic,
		msg:     msg,
	}
}
func (c *Context) Topic() string {
	return c.topic
}
func (c *Context) Msg() interface{} {
	return c.msg
}
