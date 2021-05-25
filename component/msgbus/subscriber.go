/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package msgbus

import (
	"runtime"
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

// 全局自增订阅者id
var autoIncrSubscriberId uint32

// 生成下一个订阅者id
func nextSubscriberId() uint32 {
	return atomic.AddUint32(&autoIncrSubscriberId, 1)
}

// 订阅者
type subscriber struct {
	handler core.MsgbusHandler
	queue   chan *channelMsg

	isClose uint32
}

func newSubscriber(queueSize, threadCount int, handler core.MsgbusHandler) *subscriber {
	if queueSize < 1 {
		queueSize = consts.DefaultMsgbusQueueSize
	}

	if threadCount < 1 {
		threadCount = runtime.NumCPU() >> 1
		if threadCount < 1 {
			threadCount = 1
		}
	}
	// 创建订阅者
	sub := &subscriber{
		handler: handler,
		queue:   make(chan *channelMsg, queueSize),
	}

	// 开始消费
	for i := 0; i < threadCount; i++ {
		go sub.start()
	}

	return sub
}

func (s *subscriber) start() {
	for msg := range s.queue {
		s.process(msg)
	}
}

func (s *subscriber) process(msg *channelMsg) {
	ctx := newContext(msg)

	ctx.Debug("msgbus.receive")

	err := utils.Recover.WrapCall(func() error {
		return s.handler(ctx)
	})

	if err == nil {
		ctx.Debug("msgbus.success")
		return
	}

	ctx.Error("msgbus.error!", zap.String("error", utils.Recover.GetRecoverErrorDetail(err)))
}

// 关闭
func (s *subscriber) Close() {
	if atomic.CompareAndSwapUint32(&s.isClose, 0, 1) {
		close(s.queue)
	}
}
