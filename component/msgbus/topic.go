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
type msgTopic struct {
	subs map[uint32]*subscriber
	mx   sync.RWMutex // 用于锁 subs
}

func newMsgTopic() *msgTopic {
	return &msgTopic{
		subs: make(map[uint32]*subscriber),
	}
}

// 发布消息
func (t *msgTopic) Publish(topic string, msg interface{}) {
	t.mx.RLock()
	for _, sub := range t.subs {
		sub.queue <- &channelMsg{
			topic: topic,
			msg:   msg,
		}
	}
	t.mx.RUnlock()
}

// 订阅
func (t *msgTopic) Subscribe(threadCount int, handler core.MsgbusHandler) (subscriberId uint32) {
	sub := newSubscriber(threadCount, handler)
	subscriberId = nextSubscriberId()

	t.mx.Lock()
	t.subs[subscriberId] = sub
	t.mx.Unlock()

	return subscriberId
}

// 取消订阅, 会将订阅者关闭
func (t *msgTopic) Unsubscribe(subscribeId uint32) {
	t.mx.Lock()
	sub, ok := t.subs[subscribeId]
	if ok {
		sub.Close()
		delete(t.subs, subscribeId)
	}
	t.mx.Unlock()
}

// 关闭主题
func (t *msgTopic) Close() {
	t.mx.Lock()
	for _, sub := range t.subs {
		sub.Close()
	}

	// 如果不清除, 在调用 Publish 会导致panic
	t.subs = make(map[uint32]*subscriber)

	t.mx.Unlock()
}
