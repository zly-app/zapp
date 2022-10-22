package main

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config/watch_example/example_provider"
)

// 可以在定义变量时初始化
var MyConfigWatch = zapp.WatchConfigKey("group_name", "key_name")

func main() {
	app := zapp.NewApp("test",
		example_provider.WithPlugin(true), // 启用插件并设为默认提供者
	)
	defer app.Exit()

	// 也可以在这里初始化
	//MyConfigWatch = zapp.WatchConfigKey("group_name", "key_name")

	// 获取原始数据
	y1 := MyConfigWatch.GetString()
	app.Info(y1) // 1

	// 转为 int 值
	y2 := MyConfigWatch.GetInt()
	app.Info(y2) // 1

	// 转为 boolean 值
	y3 := MyConfigWatch.GetBool()
	app.Info(y3) // true

	// 检查复合预期
	b1 := MyConfigWatch.Expect("1")
	b2 := MyConfigWatch.Expect(1)
	b3 := MyConfigWatch.Expect(true)
	app.Info(b1, b2, b3) // true, true, true
}
