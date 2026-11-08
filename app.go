package zapp

import (
	"context"
	"os"
	"sync"

	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/component"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/daemon"
	"github.com/zly-app/zapp/freemem"
	"github.com/zly-app/zapp/lifecycle"
	"github.com/zly-app/zapp/log"
	"github.com/zly-app/zapp/plugin"
	"github.com/zly-app/zapp/service"
)

var defaultApp core.IApp

// App 返回最近一次 NewApp 创建的应用
//
// 注意：这是进程级单例，属于框架便利入口；模型之间不要用它互相取依赖，
// 需要什么就在构造时注入什么。
func App() core.IApp {
	return defaultApp
}

// App 是应用壳（Shell）：只做三件事 —— 装配模型、持有基础上下文、把 IApp 的方法委托给模型。
//
// 壳里不写任何业务规则与生命周期时序：时序在 lifecycle（业务层），
// 数据与规则在各模型（plugin / service / component / daemon / freemem）。
// 允许调用：全部模型 + lifecycle（组合根的特权）。禁止：在壳里长出业务判断。
type shell struct {
	name string
	opt  *option

	baseCtx       context.Context
	baseCtxCancel context.CancelFunc

	config core.IConfig
	core.ILogger

	models  lifecycle.Models
	options lifecycle.Options

	onceExit sync.Once
}

var _ core.IApp = (*shell)(nil)

// NewApp 创建并初始化一个应用
//
// 初始化时序由 lifecycle.Initialize 负责；本函数只做装配与编排入口。
func NewApp(appName string, opts ...Option) core.IApp {
	if appName == "" {
		log.Fatal("appName is empty")
	}

	opt := newOption(opts...)
	app := &shell{name: appName, opt: opt}
	app.baseCtx, app.baseCtxCancel = context.WithCancel(context.Background())
	app.models.Component = component.NewContainer()
	defaultApp = app

	// 守护进程：参数处理必须发生在配置加载之前（install/stop 等命令不依赖配置，
	// 且配置缺失时仍要能执行），因此先于 config/logger 构建。
	app.models.Daemon = daemon.NewManager(appName, opt.EnableDaemon, app.run, app.Exit)
	app.models.Daemon.Prepare()

	// 配置模型
	app.config = opt.Config
	if app.config == nil {
		app.config = config.NewConfig(appName, opt.ConfigOpts...)
	}

	// 日志模型
	app.ILogger = log.NewLogger(appName, app.config, opt.LogOpts...)

	// 其余模型：都依赖 app 与 logger，因此必须在上面两步之后构建
	app.models.Plugin = plugin.NewManager(app, app.ILogger)
	app.models.Service = service.NewManager(app, app.ILogger, opt.ExitOnErrOfObserveServiceUnstable)
	app.models.FreeMem = freemem.New(app.config.Config().Frame.FreeMemoryInterval)

	app.options = lifecycle.Options{
		Plugins:              opt.Plugins,
		CustomEnablePlugins:  opt.CustomEnablePluginsFn,
		Services:             opt.Services,
		CustomEnableServices: opt.CustomEnableServicesFn,
		CustomComponent:      opt.CustomComponentFn,
		Handlers:             opt.Handlers,
	}

	lifecycle.Initialize(app, app.models, app.options)
	return app
}

// run 启动用例（由守护进程拉起或 Run 驱动）
func (a *shell) run() {
	lifecycle.Start(a, a.models, a.options)
}

// exit 退出用例
func (a *shell) exit() {
	lifecycle.Exit(a, a.models, a.baseCtxCancel, a.options)
}

func (a *shell) Name() string {
	return a.name
}

// Run 启动应用并阻塞
func (a *shell) Run() {
	if err := a.models.Daemon.Run(); err != nil {
		log.Fatal("Run err", zap.Error(err))
	}
}

// Exit 退出应用
func (a *shell) Exit() {
	a.onceExit.Do(func() {
		a.exit()
		os.Exit(0)
	})
}

func (a *shell) BaseContext() context.Context {
	return a.baseCtx
}

func (a *shell) GetConfig() core.IConfig {
	return a.config
}

func (a *shell) GetLogger() core.ILogger {
	return a.ILogger
}

func (a *shell) GetComponent() core.IComponent {
	return a.models.Component.Get()
}

func (a *shell) GetPlugin(pluginType core.PluginType) (core.IPlugin, bool) {
	return a.models.Plugin.Get(pluginType)
}

func (a *shell) InjectPlugin(pluginType core.PluginType, v ...interface{}) {
	a.models.Plugin.Inject(pluginType, a.opt.IgnoreInjectOfDisablePlugin, v...)
}

func (a *shell) GetService(serviceType core.ServiceType) (core.IService, bool) {
	return a.models.Service.Get(serviceType)
}

func (a *shell) InjectService(serviceType core.ServiceType, v ...interface{}) {
	a.models.Service.Inject(serviceType, a.opt.IgnoreInjectOfDisableService, v...)
}
