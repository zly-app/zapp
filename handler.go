/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/8/22
   Description :
-------------------------------------------------
*/

package zapp

import (
	"github.com/zly-app/zapp/core"
)

type Handler func(app core.IApp, handlerType HandlerType)

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
