package lifecycle

import (
	"context"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/zly-app/zapp/component"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/daemon"
	"github.com/zly-app/zapp/filter"
	"github.com/zly-app/zapp/freemem"
	"github.com/zly-app/zapp/handler"
	"github.com/zly-app/zapp/plugin"
	"github.com/zly-app/zapp/service"
)

// Models 是业务层编排的模型集合（组合根注入的 port 组）
//
// 业务层只认这些窄接口背后的操作，不认识它们的内部数据。
type Models struct {
	Plugin    *plugin.Manager
	Service   *service.Manager
	Component *component.Container
	Daemon    *daemon.Manager
	FreeMem   *freemem.Releaser
}

// Options 是编排所需的启用清单与钩子
//
// 这些是「本次装配的选择」，不是模型数据，因此不进模型、只由业务层持有。
type Options struct {
	Plugins              []core.PluginType
	CustomEnablePlugins  []func(core.IApp, []core.PluginType) []core.PluginType
	Services             []core.ServiceType
	CustomEnableServices []func(core.IApp, []core.ServiceType) []core.ServiceType
	CustomComponent      []func(core.IApp, core.IComponent) core.IComponent

	Handlers map[handler.HandlerType][]handler.Handler
}

// fire 触发一个生命周期钩子
//
// 钩子是「编排事件」，所以只允许在业务层触发 —— 模型内部不得自行触发钩子。
func fire(app core.IApp, o Options, t handler.HandlerType) {
	handler.Trigger(app, t)
	for _, h := range o.Handlers[t] {
		h(app, t)
	}
}

// Initialize 是「应用初始化」用例
//
// 顺序即契约：配置打印 → 组件 → 插件 → 过滤器 → 服务，每步前后触发对应钩子。
// 任一步不可恢复的失败由该步的模型报告，业务层负责终止。
func Initialize(app core.IApp, m Models, o Options) {
	printConfig(app)

	fire(app, o, handler.BeforeInitializeHandler)
	app.Info("app初始化")

	fire(app, o, handler.BeforeMakeComponent)
	m.Component.Make(app, o.CustomComponent)
	fire(app, o, handler.AfterMakeComponent)

	fire(app, o, handler.BeforeMakePlugin)
	if err := m.Plugin.Make(o.Plugins, o.CustomEnablePlugins); err != nil {
		app.Fatal("插件检查失败", zap.Error(err))
	}
	fire(app, o, handler.AfterMakePlugin)

	fire(app, o, handler.BeforeMakeFilter)
	filter.MakeFilter()
	filter.InitFilter(app)
	fire(app, o, handler.AfterMakeFilter)

	fire(app, o, handler.BeforeMakeService)
	if err := m.Service.Make(o.Services, o.CustomEnableServices); err != nil {
		app.Fatal("服务检查失败", zap.Error(err))
	}
	fire(app, o, handler.AfterMakeService)

	app.Info("app初始化完毕")
	fire(app, o, handler.AfterInitializeHandler)
}

// Start 是「应用启动」用例
//
// 顺序即契约：插件（按依赖）→ 服务（等就绪）→ 后台内存释放。
func Start(app core.IApp, m Models, o Options) {
	fire(app, o, handler.BeforeStartHandler)
	app.Info("启动app")

	fire(app, o, handler.BeforeStartPlugin)
	if err := m.Plugin.Start(); err != nil {
		app.Fatal("插件启动失败", zap.Error(err))
	}
	fire(app, o, handler.AfterStartPlugin)

	fire(app, o, handler.BeforeStartService)
	if err := m.Service.Start(); err != nil {
		app.Fatal("服务启动失败", zap.Error(err))
	}
	fire(app, o, handler.AfterStartService)

	m.FreeMem.Start(app.BaseContext())

	app.Info("app已启动")
	fire(app, o, handler.AfterStartHandler)
}

// Exit 是「应用退出」用例
//
// 顺序与启动相反：先停基础上下文，再服务 → 过滤器 → 插件 → 组件。
// cancel 由壳提供（它拥有基础上下文）。
func Exit(app core.IApp, m Models, cancel context.CancelFunc, o Options) {
	fire(app, o, handler.BeforeExitHandler)
	app.Info("app准备退出")

	cancel()

	fire(app, o, handler.BeforeCloseService)
	m.Service.Close()
	fire(app, o, handler.AfterCloseService)

	fire(app, o, handler.BeforeCloseFilter)
	filter.CloseFilter()
	fire(app, o, handler.AfterCloseFilter)

	fire(app, o, handler.BeforeClosePlugin)
	m.Plugin.Close()
	fire(app, o, handler.AfterClosePlugin)

	fire(app, o, handler.BeforeCloseComponent)
	m.Component.Close()
	fire(app, o, handler.AfterCloseComponent)

	app.Warn("app已退出")
	fire(app, o, handler.AfterExitHandler)
}

// printConfig 按配置决定是否把完整配置打印到日志
func printConfig(app core.IApp) {
	cfg := app.GetConfig()
	if cfg == nil || !cfg.Config().Frame.PrintConfig {
		return
	}
	data, err := yaml.Marshal(cfg.GetViper().AllSettings())
	if err != nil {
		app.Error("打印配置时序列化失败", zap.Error(err))
		return
	}
	app.Info("配置数据:\n", string(data), "\n")
}
