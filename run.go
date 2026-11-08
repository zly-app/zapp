package zapp

import (
	"time"

	"github.com/zly-app/zapp/log"
)

// Run 是一个前台常驻示例入口
//
// 保留供示例程序直接调用；它不参与应用初始化时序，与 App.Run 无关。
func Run() {
	log.Info("Service start...")
	// 在这里编写你的服务逻辑
	for {
		log.Info("Service is running...")
		time.Sleep(1 * time.Second)
	}
}
