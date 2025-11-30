/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/3/19
   Description :
-------------------------------------------------
*/

package msgbus

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/zly-app/zapp/core"
)

func TestTopic(t *testing.T) {
	const loopCount = 3

	var wg sync.WaitGroup
	wg.Add(loopCount * 2)

	topic1 := newMsgTopic()
	defer topic1.Close()

	topic2 := newMsgTopic()
	defer topic2.Close()

	subscribe1 := topic1.Subscribe(10, 0, func(ctx context.Context, msg core.IMsgbusMessage) {
		fmt.Println("subscribe.topic1", msg)
		wg.Done()
	})

	subscribe2 := topic2.Subscribe(10, 0, func(ctx context.Context, msg core.IMsgbusMessage) {
		fmt.Println("subscribe.topic2", msg)
		wg.Done()
	})

	for i := 0; i < loopCount; i++ {
		topic1.Publish(context.Background(), "topic1", i)
		topic2.Publish(context.Background(), "topic2", i)
	}

	topic1.Unsubscribe(subscribe1)
	topic2.Unsubscribe(subscribe2)

	wg.Wait()
}
