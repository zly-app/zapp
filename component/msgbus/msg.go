package msgbus

import (
	"context"

	"github.com/zly-app/zapp/core"
)

// 通道消息
type channelMsg struct {
	ctx   context.Context
	topic string
	msg   interface{}
}

func newMessage(ctx context.Context, topic string, msg interface{}) core.IMsgbusMessage {
	return &channelMsg{
		ctx:   ctx,
		topic: topic,
		msg:   msg,
	}
}

func (m *channelMsg) Ctx() context.Context { return m.ctx }
func (m *channelMsg) Topic() string        { return m.topic }
func (m *channelMsg) Msg() interface{}     { return m.msg }

type msgDataFlag struct{}

func saveMsgData(ctx context.Context, msgData string) context.Context {
	return context.WithValue(ctx, msgDataFlag{}, msgData)
}
func getMsgData(ctx context.Context) (string, bool) {
	v := ctx.Value(msgDataFlag{})
	if v == nil {
		return "", false
	}

	msgData, ok := v.(string)
	return msgData, ok
}
