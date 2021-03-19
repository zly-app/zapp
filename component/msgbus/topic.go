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

// 主题
type topic struct {
	name string
	subs map[uint32]core.IMsgbusSubscriber
	mx   sync.RWMutex
}

func newMsgTopic(name string) core.IMsgbusTopic {
	return &topic{
		name: name,
		subs: make(map[uint32]core.IMsgbusSubscriber),
	}
}

func (t *topic) Name() string {
	return t.name
}

// 发布消息
func (t *topic) Publish(msg interface{}) {
	t.mx.RLock()
	for _, sub := range t.subs {
		sub.Receive(msg)
	}
	t.mx.RUnlock()
}

// 订阅
func (t *topic) Subscribe(queueSize, threadCount int, fn core.MsgbusProcessFunc) core.IMsgbusSubscriber {
	sub := newSubscriber(t.name, queueSize, threadCount, fn)

	t.mx.Lock()
	t.subs[sub.ID()] = sub
	t.mx.Unlock()

	return sub
}

// 取消订阅, 会将订阅者关闭
func (t *topic) UnsubscribeByID(subscribeId uint32) {
	t.mx.Lock()
	sub, ok := t.subs[subscribeId]
	if ok {
		delete(t.subs, subscribeId)
	}
	t.mx.Unlock()

	if ok {
		sub.Close()
	}
}

// 取消订阅, 会将订阅者关闭
func (t *topic) Unsubscribe(sub core.IMsgbusSubscriber) {
	t.mx.Lock()
	delete(t.subs, sub.ID())
	t.mx.Unlock()
	sub.Close()
}

// 关闭主题
func (t *topic) Close() {
	t.mx.Lock()
	subs := t.subs
	t.subs = make(map[uint32]core.IMsgbusSubscriber)
	t.mx.Unlock()

	for _, sub := range subs {
		sub.Close()
	}
}
