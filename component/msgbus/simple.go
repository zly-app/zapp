package msgbus

import (
	"context"

	"github.com/zly-app/zapp/core"
)

var defaultMsgBus = NewMsgBus()

func GetMsgbus() core.IMsgbus { return defaultMsgBus }

// 发布
func Publish(ctx context.Context, topic string, msg interface{}) {
	defaultMsgBus.Publish(ctx, topic, msg)
}

// 订阅, 返回订阅号
func Subscribe(topic string, threadCount int, handler core.MsgbusHandler) (subscribeId uint32) {
	return defaultMsgBus.Subscribe(topic, threadCount, handler)
}

// 全局订阅, 会收到所有消息, 返回订阅号
func SubscribeGlobal(threadCount int, handler core.MsgbusHandler) (subscribeId uint32) {
	return defaultMsgBus.SubscribeGlobal(threadCount, handler)
}

// 取消订阅
func Unsubscribe(topic string, subscribeId uint32) {
	defaultMsgBus.Unsubscribe(topic, subscribeId)
}

// 取消全局订阅
func UnsubscribeGlobal(subscribeId uint32) {
	defaultMsgBus.UnsubscribeGlobal(subscribeId)
}

// 关闭主题, 同时关闭所有订阅该主题的订阅者
func CloseTopic(topic string) {
	defaultMsgBus.CloseTopic(topic)
}

// 关闭
func Close() {
	defaultMsgBus.Close()
}
