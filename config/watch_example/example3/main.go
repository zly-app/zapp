package main

import (
	"time"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config/watch_example/example_provider"
)

type MyConfig struct {
	A int `json:"a"`
}

// 可以在定义变量时初始化
var MyConfigWatch = zapp.WatchConfigJson[*MyConfig]("group_name", "generic_key")

func main() {
	app := zapp.NewApp("test",
		example_provider.WithPlugin(true), // 启用插件并设为默认提供者
	)
	defer app.Exit()

	// 也可以在这里初始化
	//MyConfigWatch = zapp.WatchConfigJson[*MyConfig]("group_name", "generic_key")

	// 获取数据
	for {
		a := MyConfigWatch.Get()
		app.Info("数据", a)
		time.Sleep(time.Second)
	}
}
