/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/7/2
   Description :
-------------------------------------------------
*/

package core

// 组件, 如db, rpc, cache, mq等
type IComponent interface {
	// 获取app
	App() IApp
	// 获取配置
	Config() *Config

	// 日志
	ILogger
	// 关闭所有组件
	Close()

	// 协程池
	IGPools
	// 消息总线
	IMsgbus
}

// 组件类型
type ComponentType string
