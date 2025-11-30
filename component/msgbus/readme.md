# 消息总线组件

> 进程内的发布订阅

## 示例

```go
package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/zly-app/zapp/component/msgbus"
	"github.com/zly-app/zapp/core"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2) // global两次

	msgbus.Publish(context.Background(), "topic1", "msg")

	msgbus.Subscribe("topic1", 0, func(ctx context.Context, msg core.IMsgbusMessage) {
		fmt.Println("Subscribe.topic1", msg.Topic(), msg)
		wg.Done()
	})

	msgbus.Subscribe("topic2", 0, func(ctx context.Context, msg core.IMsgbusMessage) {
		fmt.Println("Subscribe.topic2", msg.Topic(), msg)
		wg.Done()
	})

	// 全局订阅
	msgbus.SubscribeGlobal(0, func(ctx context.Context, msg core.IMsgbusMessage) {
		fmt.Println("SubscribeGlobal", msg.Topic(), msg)
		wg.Done()
	})

	msgbus.Publish(context.Background(), "topic1", "msg")
	msgbus.Publish(context.Background(), "topic2", "msg")

	wg.Wait()
}
```

## 说明

+ 创建订阅者会自动创建主题
+ 发布消息时, 如果没有对应的订阅者, 则会被抛弃掉
+ 新增的订阅者不能收到历史消息
