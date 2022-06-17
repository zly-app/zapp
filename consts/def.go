/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/20
   Description :
-------------------------------------------------
*/

package consts

// 框架
const (
	// 默认清理内存间隔时间(毫秒)
	DefaultFreeMemoryInterval int = 120000
	// 默认等待服务启动阶段, 等待时间(毫秒)
	DefaultWaitServiceRunTime int = 1000
	// 默认服务不稳定观察时间, 等待时间(毫秒)
	DefaultServiceUnstableObserveTime int = 10000
)

// 配置
const (
	// 默认配置文件优先级
	DefaultConfigFiles = "./configs/default.toml,./configs/default.yaml,./configs/default.json"
	// apollo配置key
	ApolloConfigKey = "apollo"
	// 从环境变量中加载apollo集群名
	ApolloConfigClusterFromEnvKey = "ApolloCluster"
	// 包含配置文件key
	IncludeConfigFileKey = "include"
)

// 默认组件名
const DefaultComponentName = "default"

// 组件
const (
	// msgbus 默认队列大小
	DefaultMsgbusQueueSize = 1000
)
