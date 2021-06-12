/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/7/2
   Description :
-------------------------------------------------
*/

package zapp

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
)

type Option func(opt *option)

type option struct {
	// 配置选项
	ConfigOpts []config.Option
	// 日志选项
	LogOpts []zap.Option

	// 启用守护
	EnableDaemon bool
	// handlers
	Handlers map[HandlerType][]Handler
	// 在服务不稳定观察阶段中出现错误则退出
	ExitOnErrOfObserveServiceUnstable bool

	// 忽略未启用的服务注入
	IgnoreInjectOfDisableServer bool
	// 服务
	Services map[core.ServiceType]bool
	// 服务选项
	ServicesOpts map[core.ServiceType][]interface{}
	// 自定义服务函数
	CustomServicesFn func(app core.IApp, services map[core.ServiceType]bool, servicesOpts map[core.ServiceType][]interface{})

	// 自定义组件函数
	CustomComponentFn func(app core.IApp) core.IComponent
}

func newOption(opts ...Option) *option {
	opt := &option{
		EnableDaemon:                      false,
		Handlers:                          make(map[HandlerType][]Handler),
		ExitOnErrOfObserveServiceUnstable: true,

		IgnoreInjectOfDisableServer: false,
		Services:                    make(map[core.ServiceType]bool),
		ServicesOpts:                make(map[core.ServiceType][]interface{}),
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

// 检查自定义启用服务
func (o *option) CheckCustomServices(app core.IApp) {
	if o.CustomServicesFn != nil {
		o.CustomServicesFn(app, o.Services, o.ServicesOpts)
	}
}

// 设置config选项
func WithConfigOption(opts ...config.Option) Option {
	return func(opt *option) {
		opt.ConfigOpts = append(opt.ConfigOpts, opts...)
	}
}

// 日志选项
func WithLoggerOptions(opts ...zap.Option) Option {
	return func(opt *option) {
		opt.LogOpts = append(opt.LogOpts, opts...)
	}
}

// 启用守护进程模块
func WithEnableDaemon() Option {
	return func(opt *option) {
		opt.EnableDaemon = true
	}
}

// 添加handler
func WithHandler(t HandlerType, hs ...Handler) Option {
	return func(opt *option) {
		opt.Handlers[t] = append(opt.Handlers[t], hs...)
	}
}

// 在服务不稳定观察阶段中出现错误则退出, 默认true
func WithExitOnErrOfObserveServiceUnstable(exit ...bool) Option {
	return func(opt *option) {
		opt.ExitOnErrOfObserveServiceUnstable = len(exit) == 0 || exit[0]
	}
}

// 忽略未启用的服务注入
func WithIgnoreInjectOfDisableServer(ignore ...bool) Option {
	return func(opt *option) {
		opt.IgnoreInjectOfDisableServer = len(ignore) == 0 || ignore[0]
	}
}

// 启动服务
func WithService(serviceType core.ServiceType, enable ...bool) Option {
	return func(opt *option) {
		opt.Services[serviceType] = len(enable) == 0 || enable[0]
	}
}

// 服务选项
func WithServiceOpts(serviceType core.ServiceType, opts ...interface{}) Option {
	return func(opt *option) {
		opt.ServicesOpts[serviceType] = append(opt.ServicesOpts[serviceType], opts...)
	}
}

// 自定义服务
//
// 与WithService不同的是这里已经加载了component, 用户可以方便的根据各种条件启用和关闭服务.
// 示例:
// 		zapp.WithCustomService(func(app core.IApp, services map[core.ServiceType]bool, servicesOpts map[core.ServiceType][]interface{}) {
// 			services["api"] = true
// 		}),
func WithCustomService(fn func(app core.IApp, services map[core.ServiceType]bool, servicesOpts map[core.ServiceType][]interface{})) Option {
	return func(opt *option) {
		opt.CustomServicesFn = fn
	}
}

// 自定义组件
func WithCustomComponent(creator func(app core.IApp) core.IComponent) Option {
	return func(opt *option) {
		opt.CustomComponentFn = creator
	}
}
