/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package msgbus

import (
	"strconv"
	"sync/atomic"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/log"
	"github.com/zly-app/zapp/pkg/utils"
)

type Subscriber interface {
	// 获取订阅者id
	GetSubId() uint32
	// 启动
	Start()
	// 关闭订阅者, 程序结束前注意要调用这个方法
	Close()

	Handler(msg core.IMsgbusMessage)
}

// 全局自增订阅者id
var autoIncrSubscriberId uint32

// 生成下一个订阅者id
func nextSubscriberId() uint32 {
	return atomic.AddUint32(&autoIncrSubscriberId, 1)
}

// 订阅者
type subscriber struct {
	isGlobal    bool
	subId       uint32
	handler     core.MsgbusHandler
	queue       chan core.IMsgbusMessage
	threadCount int

	onceStart int32
	onceClose int32
}

func newSubscriber(msgQueueSize int, threadCount int, handler core.MsgbusHandler) Subscriber {
	if threadCount < 1 {
		threadCount = 1
	}
	// 创建订阅者
	sub := &subscriber{
		subId:       nextSubscriberId(),
		handler:     handler,
		queue:       make(chan core.IMsgbusMessage, msgQueueSize),
		threadCount: threadCount,
	}
	return sub
}
func newGlobalSubscriber(msgQueueSize int, threadCount int, handler core.MsgbusHandler) Subscriber {
	if threadCount < 1 {
		threadCount = 1
	}
	// 创建订阅者
	sub := &subscriber{
		isGlobal:    true,
		subId:       nextSubscriberId(),
		handler:     handler,
		queue:       make(chan core.IMsgbusMessage, msgQueueSize),
		threadCount: threadCount,
	}
	return sub
}
func (s *subscriber) GetSubId() uint32 {
	return s.subId
}
func (s *subscriber) Start() {
	if atomic.AddInt32(&s.onceStart, 1) != 1 {
		return
	}

	for i := 0; i < s.threadCount; i++ {
		go s.start()
	}
}
func (s *subscriber) start() {
	for msg := range s.queue {
		s.process(msg)
	}
}

func (s *subscriber) Close() {
	if atomic.AddInt32(&s.onceClose, 1) == 1 {
		close(s.queue)
	}
}

func (s *subscriber) Handler(msg core.IMsgbusMessage) {
	s.queue <- msg
}

func (s *subscriber) process(msg core.IMsgbusMessage) {
	var spanName string
	if s.isGlobal {
		spanName = "msgbus/subsriber/global_" + strconv.Itoa(int(s.subId))
	} else {
		spanName = "msgbus/subsriber/" + strconv.Itoa(int(s.subId))
	}

	ctx, span := otel.Tracer("").Start(msg.Ctx(), spanName, trace.WithSpanKind(trace.SpanKindInternal))
	defer span.End()

	log.Debug(ctx, "msgbus.receive", log.String("topic", msg.Topic()), log.Uint32("subId", s.subId), log.Bool("isGlobal", s.isGlobal), log.Any("msg", msg.Msg()))
	msgData, _ := getMsgData(ctx)
	utils.Trace.CtxEvent(ctx, "receive", utils.OtelSpanKey("msg").String(msgData), utils.OtelSpanKey("isGlobal").Bool(s.isGlobal))

	err := utils.Recover.WrapCall(func() error {
		s.handler(msg.Ctx(), msg)
		return nil
	})

	if err == nil {
		log.Debug(ctx, "msgbus.success", log.String("topic", msg.Topic()), log.Uint32("subId", s.subId), log.Bool("isGlobal", s.isGlobal))
		utils.Trace.CtxEvent(ctx, "handler success")
		return
	}

	log.Error(ctx, "msgbus.error", log.String("topic", msg.Topic()), log.Uint32("subId", s.subId), log.Bool("isGlobal", s.isGlobal), log.String("error", utils.Recover.GetRecoverErrorDetail(err)))
	utils.Trace.CtxErrEvent(ctx, "handler error", err)
}
