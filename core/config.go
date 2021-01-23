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
)

// 配置结构
type Config struct {
	// 框架配置
	Frame FrameConfig

	// 服务配置
	Services ServicesConfig

	// 组件配置
	Components ComponentsConfig
}

// 配置
type IConfig interface {
	// 获取配置
	Config() *Config
	// 获取配置viper结构
	GetViper() *viper.Viper
	// 解析指定key数据到结构中
	Parse(key string, outPtr interface{}) error
	// 解析组件配置
	ParseComponentConfig(componentType ComponentType, componentName string, outPtr interface{}) error
	// 解析服务配置
	ParseServiceConfig(serviceType ServiceType, outPtr interface{}) error

	// 获取标签的值, 标签名是忽略大小写的
	GetLabel(name string) string
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
	// 标签, 注意: 标签名是忽略大小写的
	Labels map[string]string
	// log配置
	Log LogConfig
}

type LogConfig struct {
	Level                      string // 日志等级, debug, info, warn, error, dpanic, panic, fatal
	JsonEncoder                bool   // 启用json编码器, 输出的每一行日志转为json格式
	WriteToStream              bool   // 输出到屏幕
	WriteToFile                bool   // 日志是否输出到文件
	Name                       string // 日志文件名, 末尾会自动附加 .log 后缀
	AppendPid                  bool   // 是否在日志文件名后附加进程号
	Path                       string // 默认日志存放路径
	FileMaxSize                int    // 每个日志最大尺寸,单位M
	FileMaxBackupsNum          int    // 日志文件最多保存多少个备份
	FileMaxDurableTime         int    // 文件最多保存多长时间,单位天
	TimeFormat                 string // 时间显示格式
	IsTerminal                 bool   // 是否为控制台模式(控制台会打印彩色日志等级)
	DevelopmentMode            bool   // 开发者模式, 在开发者模式下日志记录器在写完DPanic消息后程序会感到恐慌
	ShowFileAndLinenum         bool   // 显示文件路径和行号
	ShowFileAndLinenumMinLevel string // 最小显示文件路径和行号的等级
	MillisDuration             bool   // 对zap.Duration转为毫秒
}

// 服务配置
type ServicesConfig map[string]interface{}

// 组件配置
type ComponentsConfig map[string]map[string]interface{}
