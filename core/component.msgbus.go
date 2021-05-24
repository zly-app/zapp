/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package core

// 消息总线
type IMsgbus interface {
	// 获取topic, 如果topic不存在自动创建一个
	GetMsgbusTopic(name string) IMsgbusTopic
	// 关闭topic
	CloseMsgbusTopic(name string)
	// 关闭
	Close()
}

// 消息总线的主题
type IMsgbusTopic interface {
	// 话题名
	Name() string
	// 发布消息
	Publish(msg interface{})
	// 订阅
	Subscribe(queueSize, threadCount int, fn MsgbusProcessFunc) IMsgbusSubscriber
	// 取消订阅, 会将订阅者关闭
	UnsubscribeByID(subscribeId uint32)
	// 取消订阅, 会将订阅者关闭
	Unsubscribe(subscribe IMsgbusSubscriber)
	// 关闭主题
	Close()
}

// 消息总线的订阅者
type IMsgbusSubscriber interface {
	// 订阅者id
	ID() uint32
	// 话题名
	TopicName() string
	// 接收消息
	Receive(msg interface{})
	// 关闭, 关闭后禁止调用 Receive 方法, 否则可能会产生panic
	Close()
}

type IMsgbusContext interface {
	ILogger
	Topic() string
	Msg() interface{}
}

// 处理函数
type MsgbusProcessFunc = func(ctx IMsgbusContext) error
