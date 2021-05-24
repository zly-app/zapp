/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package msgbus

import (
	"testing"
	"time"

	"github.com/zly-app/zapp/core"
)

func TestTopic(t *testing.T) {
	topic := newMsgTopic("test")
	defer topic.Close()

	subscribe := topic.Subscribe(10, 1, func(ctx core.IMsgbusContext) error {
		ctx.Info(ctx.Msg)
		return nil
	})

	for i := 0; i < 10; i++ {
		topic.Publish(i)
	}

	topic.Unsubscribe(subscribe)
	time.Sleep(time.Second)
}
