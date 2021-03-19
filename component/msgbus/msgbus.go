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
	topics map[string]core.IMsgbusTopic
	mx     sync.RWMutex
}

func NewMsgbus() core.IMsgbus {
	return &msgbus{
		topics: make(map[string]core.IMsgbusTopic),
	}
}

// 获取topic, 如果topic不存在自动创建一个
func (m *msgbus) GetMsgbusTopic(name string) core.IMsgbusTopic {
	m.mx.RLock()
	t, ok := m.topics[name]
	m.mx.RUnlock()

	if ok {
		return t
	}

	m.mx.Lock()
	t, ok = m.topics[name]
	if !ok {
		t = newMsgTopic(name)
		m.topics[name] = t
	}
	m.mx.Unlock()
	return t
}

// 关闭topic
func (m *msgbus) CloseMsgbusTopic(name string) {
	m.mx.Lock()
	t, ok := m.topics[name]
	if ok {
		delete(m.topics, name)
	}
	m.mx.Unlock()

	if ok {
		t.Close()
	}
}

// 关闭
func (m *msgbus) Close() {
	m.mx.Lock()
	topics := m.topics
	m.topics = make(map[string]core.IMsgbusTopic)
	m.mx.Unlock()

	for _, topic := range topics {
		topic.Close()
	}
}
