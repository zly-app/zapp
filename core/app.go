/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/7/2
   Description :
-------------------------------------------------
*/

package core

import (
	"context"
)

// app
//
// 用于将所有模块连起来
type IApp interface {
	// app名
	Name() string
	// 启动
	//
	// 开启所有服务并挂起
	Run()
	// 退出
	//
	// 结束所有服务并退出
	Exit()
	// 基础上下文, 这个用于监听服务结束, app会在关闭服务之前调用cancel()
	BaseContext() context.Context

	// 获取配置
	GetConfig() IConfig

	// 日志组件
	ILogger
	// 获取日志组件
	GetLogger() ILogger

	// 获取组件
	GetComponent() IComponent

	// 获取插件
	GetPlugin(pluginType PluginType) (IPlugin, bool)
	// 注入插件
	InjectPlugin(pluginType PluginType, a ...interface{})

	// 获取服务
	GetService(serviceType ServiceType) (IService, bool)
	// 注入服务
	InjectService(serviceType ServiceType, a ...interface{})
}

type Depender interface {
	/*依赖其它启动项

	plugin 在调用Start时会检查依赖项, 其依赖项名称为插件类型
	service	在调用Start时会检查依赖项, 其依赖项名称为服务类型
	*/
	DependsOn() []string
}
