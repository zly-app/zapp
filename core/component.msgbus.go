/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package core

import (
	"context"
)

// 消息总线
type IMsgbus interface {
	// 发布
	Publish(ctx context.Context, topic string, msg interface{})
	// 订阅, 返回订阅号
	Subscribe(topic string, threadCount int, handler MsgbusHandler) (subscribeId uint32)
	// 全局订阅, 会收到所有消息
	SubscribeGlobal(threadCount int, handler MsgbusHandler) (subscribeId uint32)
	// 取消订阅
	Unsubscribe(topic string, subscribeId uint32)
	// 取消全局订阅
	UnsubscribeGlobal(subscribeId uint32)
	// 关闭主题, 同时关闭所有订阅该主题的订阅者
	CloseTopic(topic string)
	// 关闭
	Close()
}

type IMsgbusMessage interface {
	Ctx() context.Context
	Topic() string
	Msg() interface{}
}

// 处理函数
type MsgbusHandler = func(ctx context.Context, msg IMsgbusMessage)
