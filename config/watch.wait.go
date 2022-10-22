package config

import (
	"sync"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/handler"
)

var watchWaitWG sync.WaitGroup

func init() {
	watchWaitWG.Add(1)
	handler.AddHandler(handler.AfterInitializeHandler, func(app core.IApp, handlerType handler.HandlerType) {
		watchWaitWG.Done()
	})
}

// 等待直到app初始化成功
func waitAppInit() {
	watchWaitWG.Wait()
}
