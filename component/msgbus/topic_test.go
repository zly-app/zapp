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
	"testing"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
)

func TestTopic(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(10)

	topic1 := newMsgTopic()
	defer topic1.Close()

	topic2 := newMsgTopic()
	defer topic2.Close()

	subscribe1 := topic1.Subscribe(1, func(ctx core.IMsgbusContext) error {
		ctx.Info("subscribe.topic1", zap.Any("msg", ctx.Msg()))
		wg.Done()
		return nil
	})

	subscribe2 := topic2.Subscribe(1, func(ctx core.IMsgbusContext) error {
		ctx.Info("subscribe.topic2", zap.Any("msg", ctx.Msg()))
		wg.Done()
		return nil
	})

	for i := 0; i < 5; i++ {
		topic1.Publish("topic1", i)
		topic2.Publish("topic2", i)
	}

	topic1.Unsubscribe(subscribe1)
	topic2.Unsubscribe(subscribe2)

	wg.Wait()
}
