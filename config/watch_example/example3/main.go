package main

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/config/watch_example/example_provider"
)

func main() {
	app := zapp.NewApp("test",
		example_provider.WithPlugin(true), // 启用插件并设为默认提供者
	)
	defer app.Exit()

	type AA struct {
		A int `json:"a"`
	}

	// 获取key对象
	keyObj := config.WatchKeyStruct[*AA]("group_name", "generic_key", config.WithWatchStructJson())
	a := keyObj.Get()
	app.Info("数据", a)

	keyObj.AddCallback(func(first bool, oldData, newData *AA) {
		app.Info("回调",
			zap.Bool("first", first),
			zap.Any("oldData", oldData),
			zap.Any("newData", newData),
		)
	})
	app.Run()
}
