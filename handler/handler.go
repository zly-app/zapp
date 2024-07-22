/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/8/22
   Description :
-------------------------------------------------
*/

package handler

import (
	"github.com/zly-app/zapp/core"
)

type Handler func(app core.IApp, handlerType HandlerType)

var handlers = map[HandlerType][]Handler{}

// handler类型
type HandlerType int

const (
	// 在app初始化前
	BeforeInitializeHandler HandlerType = iota + 1
	// 构建组件前
	BeforeMakeComponent
	// 构建组件后
	AfterMakeComponent
	// 构建插件前
	BeforeMakePlugin
	// 构建插件后
	AfterMakePlugin
	// 构建过滤器前
	BeforeMakeFilter
	// 构建过滤器后
	AfterMakeFilter
	// 构建服务前
	BeforeMakeService
	// 构建服务后
	AfterMakeService
	// 在app初始化后
	AfterInitializeHandler

	// 在app启动前
	BeforeStartHandler
	// 在启动插件前
	BeforeStartPlugin
	// 在启动插件后
	AfterStartPlugin
	// 在启动服务前
	BeforeStartService
	// 在启动服务后
	AfterStartService
	// 在app启动后
	AfterStartHandler

	// 在app退出前
	BeforeExitHandler
	// 在关闭服务前
	BeforeCloseService
	// 在关闭服务后
	AfterCloseService
	// 在关闭过滤器前
	BeforeCloseFilter
	// 在关闭过滤器后
	AfterCloseFilter
	// 在关闭插件前
	BeforeClosePlugin
	// 在关闭插件后
	AfterClosePlugin
	// 在关闭组件前
	BeforeCloseComponent
	// 在关闭组件后
	AfterCloseComponent
	// 在app退出后
	AfterExitHandler
)

// 添加handler, 和WithHandler不同的是, 它可以在NewApp之前执行, 并且它的执行顺序优先于WithHandler
func AddHandler(t HandlerType, hs ...Handler) {
	handlers[t] = append(handlers[t], hs...)
}

// 触发
func Trigger(app core.IApp, t HandlerType) {
	for _, h := range handlers[t] {
		h(app, t)
	}
}
