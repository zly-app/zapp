package metrics

import (
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

// 默认组件类型
const DefaultComponentType core.ComponentType = "metrics"

const (
	// 默认启用进程收集器
	defaultProcessCollector = true
	// 默认启用go收集器
	defaultGoCollector = true
	// 默认拉取路径`
	defaultPullPath = "/metrics"
	// 默认push模式推送时间间隔
	defaultPushTimeInterval = 10000
	// 默认push模式推送重试次数
	defaultPushRetry = 2
	// 默认push模式推送重试时间间隔
	defaultPushRetryInterval = 1000
)

type Config struct {
	ProcessCollector bool // 启用进程收集器
	GoCollector      bool // 启用go收集器

	PullBind string // pull模式bind地址, 如: ':9100', 如果为空则不启用pull模式
	PullPath string // pull模式拉取路径, 如: '/metrics'

	PushAddress string // push模式 pushGateway地址, 如果为空则不启用push模式, 如: 'http://127.0.0.1:9091'
	/*push模式 instance 标记的值
	  这个值用于区分相同服务的不同实例.
	  如果为空则设为主机名, 如果无法获取主机名则设为app名.
	*/
	PushInstance      string // 实例, 一般为ip或主机名
	PushTimeInterval  int64  // push模式推送时间间隔, 单位毫秒
	PushRetry         uint32 // push模式推送重试次数
	PushRetryInterval int64  // push模式推送重试时间间隔, 单位毫秒
}

func newConfig() *Config {
	return &Config{
		ProcessCollector: defaultProcessCollector,
		GoCollector:      defaultGoCollector,
		PushRetry:        defaultPushRetry,
	}
}

func (conf *Config) Check() {
	if conf.PullPath == "" {
		conf.PullPath = defaultPullPath
	}

	if conf.PushInstance == "" {
		conf.PushInstance = utils.GetInstance("")
	}
	if conf.PushTimeInterval < 1 {
		conf.PushTimeInterval = defaultPushTimeInterval
	}
	if conf.PushRetryInterval < 1 {
		conf.PushRetryInterval = defaultPushRetryInterval
	}
}
