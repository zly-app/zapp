/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/7/2
   Description :
-------------------------------------------------
*/

package zapp

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
)

type Option func(opt *option)

type option struct {
	// 自定义配置
	Config core.IConfig
	// 配置选项, 如果设置了自定义配置则配置选项不生效
	ConfigOpts []config.Option
	// 日志选项
	LogOpts []zap.Option

	// 启用守护
	EnableDaemon bool
	// handlers
	Handlers map[HandlerType][]Handler

	// 忽略未启用的插件注入
	IgnoreInjectOfDisablePlugin bool
	// 插件
	Plugins []core.PluginType
	// 自定义启用插件函数
	CustomEnablePluginsFn func(app core.IApp, plugins []core.PluginType) []core.PluginType

	// 在服务不稳定观察阶段中出现错误则退出
	ExitOnErrOfObserveServiceUnstable bool
	// 忽略未启用的服务注入
	IgnoreInjectOfDisableService bool
	// 服务
	Services []core.ServiceType
	// 自定义启用服务函数
	CustomEnableServicesFn func(app core.IApp, services []core.ServiceType) []core.ServiceType

	// 自定义组件函数
	CustomComponentFn func(app core.IApp) core.IComponent
}

func newOption(opts ...Option) *option {
	opt := &option{
		EnableDaemon:                      false,
		Handlers:                          make(map[HandlerType][]Handler),
		ExitOnErrOfObserveServiceUnstable: true,

		IgnoreInjectOfDisablePlugin: false,
		Plugins:                     make([]core.PluginType, 0),

		IgnoreInjectOfDisableService: false,
		Services:                     make([]core.ServiceType, 0),
	}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

// 检查自定义启用插件
func (o *option) CheckPlugins(app core.IApp) error {
	if o.CustomEnablePluginsFn != nil {
		o.Plugins = o.CustomEnablePluginsFn(app, o.Plugins)
	}

	mm := make(map[core.PluginType]struct{}, len(o.Plugins))
	for _, t := range o.Plugins {
		l := len(mm)
		mm[t] = struct{}{}
		if len(mm) == l {
			return fmt.Errorf("插件启用重复: %v", t)
		}
	}
	return nil
}

// 检查自定义启用服务
func (o *option) CheckServices(app core.IApp) error {
	if o.CustomEnableServicesFn != nil {
		o.Services = o.CustomEnableServicesFn(app, o.Services)
	}

	mm := make(map[core.ServiceType]struct{}, len(o.Services))
	for _, t := range o.Services {
		l := len(mm)
		mm[t] = struct{}{}
		if len(mm) == l {
			return fmt.Errorf("服务启用重复: %v", t)
		}
	}
	return nil
}

// 自定义配置
func WithCustomConfig(config core.IConfig) Option {
	return func(opt *option) {
		opt.Config = config
	}
}

// 设置config选项. 如果设置了自定义配置则配置选项不生效
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

// 忽略未启用的插件注入
func WithIgnoreInjectOfDisablePlugin(ignore ...bool) Option {
	return func(opt *option) {
		opt.IgnoreInjectOfDisablePlugin = len(ignore) == 0 || ignore[0]
	}
}

// 启动插件(使用者不要主动调用这个函数, 应该由plugin包装, 因为plugin的选项无法通过这个函数传递)
func WithPlugin(pluginType core.PluginType, enable ...bool) Option {
	return func(opt *option) {
		if len(enable) == 0 || enable[0] {
			opt.Plugins = append(opt.Plugins, pluginType)
		}
	}
}

// 自定义启用插件
//
// 如果要启用某个插件, 必须先注册该插件
func WithCustomEnablePlugin(fn func(app core.IApp, plugins []core.PluginType) []core.PluginType) Option {
	return func(opt *option) {
		opt.CustomEnablePluginsFn = fn
	}
}

// 在服务不稳定观察阶段中出现错误则退出, 默认true
func WithExitOnErrOfObserveServiceUnstable(exit ...bool) Option {
	return func(opt *option) {
		opt.ExitOnErrOfObserveServiceUnstable = len(exit) == 0 || exit[0]
	}
}

// 忽略未启用的服务注入
func WithIgnoreInjectOfDisableService(ignore ...bool) Option {
	return func(opt *option) {
		opt.IgnoreInjectOfDisableService = len(ignore) == 0 || ignore[0]
	}
}

// 启动服务(使用者不要主动调用这个函数, 应该由service包装, 因为service的选项无法通过这个函数传递)
func WithService(serviceType core.ServiceType, enable ...bool) Option {
	return func(opt *option) {
		if len(enable) == 0 || enable[0] {
			opt.Services = append(opt.Services, serviceType)
		}
	}
}

// 自定义启用服务
//
// 如果要启用某个服务, 必须先注册该服务
func WithCustomEnableService(fn func(app core.IApp, services []core.ServiceType) []core.ServiceType) Option {
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
