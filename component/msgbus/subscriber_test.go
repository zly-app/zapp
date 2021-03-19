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

func TestSubscriber(t *testing.T) {
	s := newSubscriber("test", 10, 1, func(ctx *core.MsgbusContext) error {
		ctx.Info(ctx.Msg)
		return nil
	})
	defer s.Close()

	for i := 0; i < 10; i++ {
		s.Receive(i)
	}

	time.Sleep(time.Second)
}
