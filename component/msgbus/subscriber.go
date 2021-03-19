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
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
)

// 全局自增订阅者id
var autoIncrSubscriberId uint32

// 生成下一个订阅者id
func nextSubscriberId() uint32 {
	return atomic.AddUint32(&autoIncrSubscriberId, 1)
}

const (
	// 默认队列大小
	defaultQueueSize = 1000
	// 默认最小队列大小
	defaultMinQueueSize = 10
	// 默认最小线程数
	defaultMinThreadCount = 1
)

// 订阅者
type subscriber struct {
	id        uint32
	topicName string
	fn        core.MsgbusProcessFunc
	queue     chan interface{}
	mx        sync.RWMutex

	isClose uint32
}

func newSubscriber(topicName string, queueSize, threadCount int, fn core.MsgbusProcessFunc) core.IMsgbusSubscriber {
	if queueSize <= 0 {
		queueSize = defaultQueueSize
	}
	if queueSize < defaultMinQueueSize {
		queueSize = defaultMinQueueSize
	}
	if threadCount < defaultMinThreadCount {
		threadCount = defaultMinThreadCount
	}

	// 创建订阅者
	sub := &subscriber{
		id:        nextSubscriberId(),
		topicName: topicName,
		fn:        fn,
		queue:     make(chan interface{}, queueSize),
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

func (s *subscriber) process(msg interface{}) {
	ctx := &core.MsgbusContext{
		ILogger: logger.Log.NewMirrorLogger(s.topicName),
		Msg:     msg,
	}

	ctx.Debug("msgbus.receive")

	err := utils.Recover.WrapCall(func() error {
		return s.fn(ctx)
	})

	if err == nil {
		ctx.Debug("msgbus.success")
		return
	}

	ctx.Error("msgbus.error!", zap.Error(err))
}

func (s *subscriber) ID() uint32 {
	return s.id
}

func (s *subscriber) TopicName() string {
	return s.topicName
}

// 接收
func (s *subscriber) Receive(msg interface{}) {
	s.queue <- msg
}

// 关闭, 关闭后禁止调用 Receive 方法, 否则可能会产生panic
func (s *subscriber) Close() {
	if atomic.CompareAndSwapUint32(&s.isClose, 0, 1) {
		close(s.queue)
	}
}
