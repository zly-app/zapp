package main

import (
	"fmt"

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

	// 获取原始数据
	y1 := MyConfigWatch.GetString()
	fmt.Println(y1) // 1

	// 转为 int 值
	y2 := MyConfigWatch.GetInt()
	fmt.Println(y2) // 1

	// 转为 boolean 值
	y3 := MyConfigWatch.GetBool()
	fmt.Println(y3) // true

	// 检查复合预期
	b1 := MyConfigWatch.Expect("1")
	b2 := MyConfigWatch.Expect(1)
	b3 := MyConfigWatch.Expect(true)
	fmt.Println(b1, b2, b3) // true, true, true
}
