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
	// 构建组件前
	BeforeMakeComponent = handler.BeforeMakeComponent
	// 构建组件后
	AfterMakeComponent = handler.AfterMakeComponent
	// 构建插件前
	BeforeMakePlugin = handler.BeforeMakePlugin
	// 构建插件后
	AfterMakePlugin = handler.AfterMakePlugin
	// 构建过滤器前
	BeforeMakeFilter = handler.BeforeMakeFilter
	// 构建过滤器后
	AfterMakeFilter = handler.AfterMakeFilter
	// 构建服务前
	BeforeMakeService = handler.BeforeMakeService
	// 构建服务后
	AfterMakeService = handler.AfterMakeService
	// 在app初始化后
	AfterInitializeHandler = handler.AfterInitializeHandler

	// 在app启动前
	BeforeStartHandler = handler.BeforeStartHandler
	// 在启动插件前
	BeforeStartPlugin = handler.BeforeStartPlugin
	// 在启动插件后
	AfterStartPlugin = handler.AfterStartPlugin
	// 在启动服务前
	BeforeStartService = handler.BeforeStartService
	// 在启动服务后
	AfterStartService = handler.AfterStartService
	// 在app启动后
	AfterStartHandler = handler.AfterStartHandler

	// 在app退出前
	BeforeExitHandler = handler.BeforeExitHandler
	// 在关闭服务前
	BeforeCloseService = handler.BeforeCloseService
	// 在关闭服务后
	AfterCloseService = handler.AfterCloseService
	// 在关闭过滤器前
	BeforeCloseFilter = handler.BeforeCloseFilter
	// 在关闭过滤器后
	AfterCloseFilter = handler.AfterCloseFilter
	// 在关闭插件前
	BeforeClosePlugin = handler.BeforeClosePlugin
	// 在关闭插件后
	AfterClosePlugin = handler.AfterClosePlugin
	// 在关闭组件前
	BeforeCloseComponent = handler.BeforeCloseComponent
	// 在关闭组件后
	AfterCloseComponent = handler.AfterCloseComponent
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
