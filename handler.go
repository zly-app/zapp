package zapp

import (
	"github.com/zly-app/zapp/handler"
)

type Handler = handler.Handler

// handler类型
type HandlerType = handler.HandlerType

const (
	// 在app初始化前
	BeforeInitializeHandler = handler.BeforeInitializeHandler
	// 在app初始化后
	AfterInitializeHandler = handler.AfterInitializeHandler
	// 在app启动前
	BeforeStartHandler = handler.BeforeStartHandler
	// 在app启动后
	AfterStartHandler = handler.AfterStartHandler
	// 在app退出前
	BeforeExitHandler = handler.BeforeExitHandler
	// 在app退出后
	AfterExitHandler = handler.AfterExitHandler
)

// 添加handler, 和WithHandler不同的是, 它可以在NewApp之前执行, 并且它的执行顺序优先于WithHandler
// 这个函数是兼容旧逻辑
func AddHandler(t HandlerType, hs ...Handler) {
	handler.AddHandler(t, hs...)
}

func (app *appCli) handler(t HandlerType) {
	handler.Trigger(app, t)
	for _, h := range app.opt.Handlers[t] {
		h(app, t)
	}
}
