/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/7/2
   Description :
-------------------------------------------------
*/

package zapp

import (
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
)

type Option func(opt *option)

type option struct {
	// 配置选项
	ConfigOpts []config.Option
	// 启用守护
	EnableDaemon bool
	// handlers
	Handlers map[HandlerType][]Handler
	// 忽略未启用的服务注入
	IgnoreInjectOfDisableServer bool
	// 服务
	Services map[core.ServiceType]bool
	// 自定义启用服务函数
	customEnableServicesFn func(c core.IComponent) (servers map[core.ServiceType]bool)
	// 自定义组件建造者
	CustomComponentCreator func(c core.IComponent) core.IComponent
}

func newOption() *option {
	return &option{
		EnableDaemon:                false,
		Handlers:                    make(map[HandlerType][]Handler),
		IgnoreInjectOfDisableServer: false,
		Services:                    make(map[core.ServiceType]bool),
	}
}

// 检查自定义启用服务
func (o *option) CheckCustomEnableServices(c core.IComponent) {
	if o.customEnableServicesFn == nil {
		return
	}

	customServices := o.customEnableServicesFn(c)
	for serviceType, enable := range customServices {
		o.Services[serviceType] = enable
	}
}

// 设置config选项
func WithConfigOption(opts ...config.Option) Option {
	return func(opt *option) {
		opt.ConfigOpts = append(opt.ConfigOpts, opts...)
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
		handlers, ok := opt.Handlers[t]
		if !ok {
			handlers = make([]Handler, 0)
		}
		handlers = append(handlers, hs...)
		opt.Handlers[t] = handlers
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

// 自定义启用哪些服务
//
// 与WithService不同的是这里已经加载了component, 用户可以方便的根据各种条件启用和关闭服务.
// 示例:
//      app.WithCustomEnableService(func(c core.IComponent) (servers map[core.ServiceType]bool) {
//			return map[core.ServiceType]bool{
//				core.CronService: true,
//			}
//		})
func WithCustomEnableService(fn func(c core.IComponent) (servers map[core.ServiceType]bool)) Option {
	return func(opt *option) {
		opt.customEnableServicesFn = fn
	}
}

// 自定义组件
func WithCustomComponent(creator func(c core.IComponent) core.IComponent) Option {
	return func(opt *option) {
		opt.CustomComponentCreator = creator
	}
}
