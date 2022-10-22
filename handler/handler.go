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
	// 在app初始化后
	AfterInitializeHandler
	// 在app启动前
	BeforeStartHandler
	// 在app启动后
	AfterStartHandler
	// 在app退出前
	BeforeExitHandler
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
