package main

import (
	"fmt"
	"time"

	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config/watch_example/example_provider"
	"github.com/zly-app/zapp/core"
)

func main() {
	app := zapp.NewApp("test",
		example_provider.WithPlugin(true), // 启用插件并设为默认提供者
	)
	defer app.Exit()

	// 获取key对象
	keyObj := app.GetConfig().WatchKey("group_name", "key_name")

	// 添加回调函数
	keyObj.AddCallback(func(w core.IConfigWatchKeyObject, first bool, oldData, newData []byte) {
		fmt.Printf("callback 回调: oldData: %s, newData: %s\n", string(oldData), string(newData))
	})

	time.Sleep(time.Second * 10)
}
