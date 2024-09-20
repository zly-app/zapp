/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/20
   Description :
-------------------------------------------------
*/

package zapp

import (
	"context"
	"os"
	"sync"

	"github.com/kardianos/service"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/depender"
)

var defaultApp core.IApp

func App() core.IApp {
	return defaultApp
}

type appCli struct {
	name string

	opt *option

	baseCtx       context.Context
	baseCtxCancel context.CancelFunc

	config core.IConfig
	core.ILogger
	component       core.IComponent
	plugins         map[core.PluginType]core.IPlugin
	pluginsDepender depender.Depender
	services        map[core.ServiceType]core.IService

	daemonService service.Service
	onceExit      sync.Once
}

// 创建一个app
//
// 根据提供的app名和选项创建一个app
// 正常启动时会初始化所有服务
func NewApp(appName string, opts ...Option) core.IApp {
	app := &appCli{
		name:     appName,
		plugins:  make(map[core.PluginType]core.IPlugin),
		services: make(map[core.ServiceType]core.IService),
		opt:      newOption(opts...),
	}
	app.baseCtx, app.baseCtxCancel = context.WithCancel(context.Background())
	defaultApp = app

	// 启用守护进程
	app.enableDaemon()

	// 配置加载
	app.config = app.opt.Config
	if app.config == nil {
		app.config = config.NewConfig(appName, app.opt.ConfigOpts...)
	}

	// app名处理
	if name := app.config.Config().Frame.Name; name != "" { // 用配置中加载的app名来替换代码传入的app名
		app.name = name
	}
	if app.name == "" {
		logger.Log.Fatal("appName is empty")
	}
	app.config.Config().Frame.Name = app.name // 配置中的app名可能是空的, 这里复写

	app.ILogger = logger.NewLogger(appName, app.config, app.opt.LogOpts...)

	if app.config.Config().Frame.PrintConfig {
		data, err := yaml.Marshal(app.config.GetViper().AllSettings())
		if err != nil {
			app.Error("打印配置时序列化失败", zap.Error(err))
		} else {
			app.Info("配置数据:\n", string(data), "\n")
		}
	}

	app.handler(BeforeInitializeHandler)
	app.Info("app初始化")

	// 构建组件
	app.makeComponent()
	// 构建插件
	app.makePlugin()
	// 构建过滤器
	app.makeFilter()
	// 构建服务
	app.makeService()

	app.Info("app初始化完毕")
	app.handler(AfterInitializeHandler)

	return app
}

func (app *appCli) run() {
	app.handler(BeforeStartHandler)
	app.Info("启动app")

	// 启动插件
	app.startPlugin()
	// 启动服务
	app.startService()
	// 开始释放内存
	app.startFreeMemory()

	app.Info("app已启动")
	app.handler(AfterStartHandler)
}

func (app *appCli) exit() {
	// app退出前
	app.handler(BeforeExitHandler)
	app.Info("app准备退出")

	// 关闭基础上下文
	app.baseCtxCancel()
	// 关闭服务
	app.closeService()
	// 关闭过滤器
	app.closeFilter()
	// 关闭插件
	app.closePlugin()
	// 释放组件资源
	app.releaseComponentResource()

	// app退出后
	app.Warn("app已退出")
	app.handler(AfterExitHandler)
}

func (app *appCli) Name() string {
	return app.name
}

// 启动
func (app *appCli) Run() {
	err := app.daemonService.Run()
	if err != nil {
		logger.Fatal("Run err", zap.Error(err))
	}
}

// 退出
func (app *appCli) Exit() {
	app.onceExit.Do(func() {
		app.exit()
		os.Exit(0)
	})
}

func (app *appCli) BaseContext() context.Context {
	return app.baseCtx
}

func (app *appCli) GetConfig() core.IConfig {
	return app.config
}

func (app *appCli) GetLogger() core.ILogger {
	return app.ILogger
}
