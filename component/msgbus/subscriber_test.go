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

	"github.com/zly-app/zapp/core"
)

func TestSubscriber(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(10)

	s := newSubscriber(10, 1, func(ctx core.IMsgbusContext) error {
		ctx.Info(ctx.Msg())
		wg.Done()
		return nil
	})
	defer s.Close()

	for i := 0; i < 10; i++ {
		s.queue <- &channelMsg{
			topic: "topic",
			msg:   i,
		}
	}

	wg.Wait()
}
