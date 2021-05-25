
# 消息总线组件

> 类似于发布订阅

## 示例

```go
package main

import (
	"sync"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"go.uber.org/zap"
)

func main() {
	app := zapp.NewApp("app1")
	defer app.Exit()

	c := app.GetComponent()

	var wg sync.WaitGroup
	wg.Add(4)

	// 订阅
	c.Subscribe("topic1", 1, func(ctx core.IMsgbusContext) error {
		c.Info("Subscribe.topic1", zap.Any("msg", ctx.Msg()))
		wg.Done()
		return nil
	})
	// 订阅
	c.Subscribe("topic2", 1, func(ctx core.IMsgbusContext) error {
		c.Info("Subscribe.topic2", zap.Any("msg", ctx.Msg()))
		wg.Done()
		return nil
	})
	// 订阅
	c.SubscribeGlobal(1, func(ctx core.IMsgbusContext) error {
		c.Info("SubscribeGlobal", zap.Any("msg", ctx.Msg()))
		wg.Done()
		return nil
	})

	c.Publish("topic1", 1)
	c.Publish("topic2", 2)

	wg.Wait()
}
```

## 说明

+ 创建订阅者会自动创建主题
+ 发布消息时, 如果没有对应的订阅者, 则会被抛弃掉
+ 订阅者不能收到历史消息.
