/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package msgbus

import (
	"sync"

	"github.com/zly-app/zapp/core"
)

// 消息总线
type msgbus struct {
	global    *msgTopic // 用于接收全局消息
	topics    map[string]*msgTopic
	mx        sync.RWMutex // 用于锁 topics
}

func (m *msgbus) Publish(topic string, msg interface{}) {
	m.global.Publish(topic, msg) // 发送消息到全局

	m.mx.RLock()
	t, ok := m.topics[topic]
	m.mx.RUnlock()

	if ok {
		t.Publish(topic, msg)
	}
}

func (m *msgbus) Subscribe(topic string, threadCount int, handler core.MsgbusHandler) (subscribeId uint32) {
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
	return t.Subscribe(threadCount, handler)
}
func (m *msgbus) SubscribeGlobal(threadCount int, handler core.MsgbusHandler) (subscribeId uint32) {
	return m.global.Subscribe(threadCount, handler)
}

func (m *msgbus) Unsubscribe(topic string, subscribeId uint32) {
	m.mx.RLock()
	t, ok := m.topics[topic]
	m.mx.RUnlock()

	if ok {
		t.Unsubscribe(subscribeId)
	}
}
func (m *msgbus) UnsubscribeGlobal(subscribeId uint32) {
	m.global.Unsubscribe(subscribeId)
}

func (m *msgbus) CloseTopic(topic string) {
	m.mx.Lock()
	t, ok := m.topics[topic]
	if ok {
		delete(m.topics, topic)
	}
	m.mx.Unlock()

	if ok {
		t.Close()
	}
}
func (m *msgbus) Close() {
	m.mx.Lock()
	m.global.Close()
	m.global = newMsgTopic()
	for _, t := range m.topics {
		t.Close()
	}
	m.topics = make(map[string]*msgTopic)
	m.mx.Unlock()
}

func NewMsgbus() core.IMsgbus {
	return &msgbus{
		global: newMsgTopic(),
		topics: make(map[string]*msgTopic),
	}
}
