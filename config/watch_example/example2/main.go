package main

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/config/watch_example/example_provider"
)

// 可以在定义变量时初始化
var MyConfigWatch = config.WatchKey("group_name", "key_name")

func main() {
	app := zapp.NewApp("test",
		example_provider.WithPlugin(true), // 启用插件并设为默认提供者
	)
	defer app.Exit()

	// 也可以在这里初始化
	//MyConfigWatch = config.WatchKey("group_name", "key_name")

	// 添加回调函数
	MyConfigWatch.AddCallback(func(first bool, oldData, newData []byte) {
		app.Info("回调",
			zap.Bool("first", first),
			zap.String("oldData", string(oldData)),
			zap.String("newData", string(newData)),
		)
	})

	app.Run()
}
