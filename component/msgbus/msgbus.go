/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package msgbus

import (
	"context"
	"sync"

	"github.com/bytedance/sonic"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/log"
	"github.com/zly-app/zapp/pkg/utils"
)

// 默认消息队列大小
const DefaultMsgQueueSize = 1000

// 消息总线
type msgBus struct {
	global       Topic // 用于接收全局消息
	topics       map[string]Topic
	msgQueueSize int
	mx           sync.RWMutex // 用于锁 topics
}

// 创建一个消息总线
func NewMsgBus() core.IMsgbus {
	return NewMsgBusWithQueueSize(DefaultMsgQueueSize)
}

// 创建一个消息总线并设置消息队列缓存大小, 消息队列满时会阻塞消息发送
func NewMsgBusWithQueueSize(msgQueueSize int) core.IMsgbus {
	if msgQueueSize < 1 {
		msgQueueSize = DefaultMsgQueueSize
	}

	return &msgBus{
		global:       newGlobalMsgTopic(),
		msgQueueSize: msgQueueSize,
		topics:       make(map[string]Topic),
	}
}

func (m *msgBus) Publish(ctx context.Context, topic string, msg interface{}) {
	ctx = utils.Ctx.CloneContext(ctx)
	ctx, span := otel.Tracer("").Start(ctx, "msgbus/publish/"+topic, trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	msgData, _ := sonic.MarshalString(msg)
	ctx = saveMsgData(ctx, msgData)
	log.Debug(ctx, "msgbus.publish", log.String("topic", topic), log.Any("msg", msgData))
	utils.Trace.CtxEvent(ctx, "publish", utils.OtelSpanKey("msg").String(msgData))

	m.global.Publish(ctx, topic, msg) // 发送消息到全局

	m.mx.RLock()
	t, ok := m.topics[topic]
	m.mx.RUnlock()

	if ok {
		t.Publish(ctx, topic, msg)
	}
}

func (m *msgBus) Subscribe(topic string, threadCount int, handler core.MsgbusHandler) (subscribeId uint32) {
	m.mx.RLock()
	t, ok := m.topics[topic]
	m.mx.RUnlock()

	if !ok {
		m.mx.Lock()
		t, ok = m.topics[topic]
		if !ok {
			t = newMsgTopic()
			m.topics[topic] = t
		}
		m.mx.Unlock()
	}
	return t.Subscribe(m.msgQueueSize, threadCount, handler)
}
func (m *msgBus) SubscribeGlobal(threadCount int, handler core.MsgbusHandler) (subscribeId uint32) {
	return m.global.Subscribe(m.msgQueueSize, threadCount, handler)
}

func (m *msgBus) Unsubscribe(topic string, subscribeId uint32) {
	m.mx.RLock()
	t, ok := m.topics[topic]
	m.mx.RUnlock()

	if ok {
		t.Unsubscribe(subscribeId)
	}
}
func (m *msgBus) UnsubscribeGlobal(subscribeId uint32) {
	m.global.Unsubscribe(subscribeId)
}

func (m *msgBus) CloseTopic(topic string) {
	m.mx.Lock()
	t, ok := m.topics[topic]
	if ok {
		delete(m.topics, topic)
	}
	m.mx.Unlock()

	// 后置关闭
	if ok {
		t.Close()
	}
}

func (m *msgBus) Close() {
	m.mx.Lock()
	clearTopic := make([]Topic, 0, 1+len(m.topics))
	clearTopic = append(clearTopic, m.global)
	for _, t := range m.topics {
		clearTopic = append(clearTopic, t)
	}
	m.global = newMsgTopic()
	m.topics = make(map[string]Topic)
	m.mx.Unlock()

	// 后置关闭
	for _, t := range clearTopic {
		t.Close()
	}
}
