/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/7/2
   Description :
-------------------------------------------------
*/

package core

import (
	"github.com/spf13/viper"
	"github.com/zlyuancn/zlog"
)

// 配置结构
type Config struct {
	// 框架配置
	Frame FrameConfig
}

// 配置
type IConfig interface {
	// 获取配置
	Config() *Config
	// 获取配置viper结构
	GetViper() *viper.Viper
	// 解析指定key数据到结构中
	Parse(key string, outPtr interface{}) error
	// 解析服务配置
	ParseServiceConfig(serviceType string, outPtr interface{}) error
	// 解析组件配置
	ParseComponentConfig(componentType, componentName string, outPtr interface{}) error

	// 获取标签的值, 标签名是忽略大小写的
	GetLabel(name string) interface{}
}

// frame配置
type FrameConfig struct {
	// debug标志
	Debug bool
	// 主动清理内存间隔时间(毫秒), <= 0 表示禁用
	FreeMemoryInterval int
	// 默认等待服务启动阶段, 等待时间(毫秒), 如果时间到未收到服务启动成功信号则将服务标记为不稳定状态然后继续开始工作(我们总不能一直等着吧)
	WaitServiceRunTime int
	// 默认服务不稳定观察时间, 等待时间(毫秒), 如果时间到仍未收到服务启动成功信号也将服务标记为启动成功
	ServiceUnstableObserveTime int
	// 标签列表, 注意: 标签名是忽略大小写的
	Labels map[string]interface{}
	// log配置
	Log LogConfig
}

type LogConfig = zlog.LogConfig
