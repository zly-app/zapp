
# msgbus 消费服务


# 示例

```go
package main

import (
	"time"

	"go.uber.org/zap"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/service/msgbus-consume"
)

func main() {
	app := zapp.NewApp("test",
		msgbus_consume.WithService(), // 启用msgbus消费服务
	)

	// 注册消费handler
	msgbus_consume.RegistryHandler(app, "test1", 1, func(ctx core.IMsgbusContext) error {
		ctx.Info("收到数据", zap.Any("msg", ctx.Msg()))
		return nil
	})

	// 发送消息
	app.GetComponent().Publish("test1", "测试消息1")
	time.Sleep(time.Second)

	app.Exit()
}
```
