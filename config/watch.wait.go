package config

import (
	"sync"
)

var watchWaitWG sync.WaitGroup

func init() {
	watchWaitWG.Add(1)
	//zapp.AddHandler(zapp.AfterInitializeHandler, func(app core.IApp, handlerType zapp.HandlerType) {
	//	watchWaitWG.Done()
	//})
}

// 等待直到app初始化成功
func waitAppInit() {
	watchWaitWG.Wait()
}
