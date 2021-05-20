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
	"os/signal"
	"sync"
	"syscall"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

type appCli struct {
	name string

	opt *option

	baseCtx       context.Context
	baseCtxCancel context.CancelFunc

	config core.IConfig
	core.ILogger
	component core.IComponent
	services  map[core.ServiceType]core.IService

	interrupt chan os.Signal
	onceExit  sync.Once
}

// 创建一个app
//
// 根据提供的app名和选项创建一个app
// 正常启动时会初始化所有服务
func NewApp(appName string, opts ...Option) core.IApp {
	if appName == "" {
		logger.Log.Fatal("appName must not empty")
	}

	app := &appCli{
		name:      appName,
		interrupt: make(chan os.Signal, 1),
		services:  make(map[core.ServiceType]core.IService),
		opt:       newOption(opts...),
	}
	app.baseCtx, app.baseCtxCancel = context.WithCancel(context.Background())

	// 启用守护进程
	app.enableDaemon()

	app.config = config.NewConfig(appName, app.opt.ConfigOpts...)
	app.ILogger = logger.NewLogger(appName, app.config, app.opt.LogOpts...)

	app.handler(BeforeInitializeHandler)
	app.Debug("app初始化")

	// 初始化组件
	app.initComponent()
	// 初始化服务
	app.initService()

	app.Debug("app初始化完毕")
	app.handler(AfterInitializeHandler)

	return app
}

func (app *appCli) run() {
	app.handler(BeforeStartHandler)
	app.Debug("启动app")

	// 启动服务
	app.startService()

	// 开始释放内存
	app.startFreeMemory()

	app.Info("app已启动")
	app.handler(AfterStartHandler)

	signal.Notify(app.interrupt, os.Kill, os.Interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-app.interrupt

	app.onceExit.Do(func() {
		app.exit()
	})
}

func (app *appCli) exit() {
	// app退出前
	app.handler(BeforeExitHandler)
	app.Debug("app准备退出")

	// 关闭基础上下文
	app.baseCtxCancel()
	// 关闭服务
	app.closeService()
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
	app.run()
}

// 退出
func (app *appCli) Exit() {
	app.onceExit.Do(func() {
		// 尝试发送退出信号
		select {
		case app.interrupt <- syscall.SIGQUIT:
		default:
		}

		app.exit()
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
