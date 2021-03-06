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
	// 自定义启用服务函数
	CustomEnableServicesFn func(app core.IApp, services map[core.ServiceType]bool)

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
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

// 检查自定义启用服务
func (o *option) CheckCustomEnableServices(app core.IApp) {
	if o.CustomEnableServicesFn != nil {
		o.CustomEnableServicesFn(app, o.Services)
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

// 启动服务(使用者不要主动调用这个函数, 应该由service包装, 因为service的选项无法通过这个函数传递)
func WithService(serviceType core.ServiceType, enable ...bool) Option {
	return func(opt *option) {
		opt.Services[serviceType] = len(enable) == 0 || enable[0]
	}
}

// 自定义服务
//
// 如果要启用某个服务, 必须使用该服务的 WithService() 选项
// 示例:
// 		zapp.WithCustomEnableService(func(app core.IApp, services map[core.ServiceType]bool) {
// 			services["api"] = true
// 		}),
func WithCustomEnableService(fn func(app core.IApp, services map[core.ServiceType]bool)) Option {
	return func(opt *option) {
		opt.CustomEnableServicesFn = fn
	}
}

// 自定义组件
func WithCustomComponent(creator func(app core.IApp) core.IComponent) Option {
	return func(opt *option) {
		opt.CustomComponentFn = creator
	}
}
