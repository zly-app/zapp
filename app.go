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
	"runtime/debug"
	"syscall"
	"time"

	"github.com/takama/daemon"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/component"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/service"
)

type appCli struct {
	name string

	opt *option

	interrupt     chan os.Signal
	baseCtx       context.Context
	baseCtxCancel context.CancelFunc

	config core.IConfig

	loggerId uint32
	core.ILogger

	component core.IComponent
	services  map[core.ServiceType]core.IService
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
		opt:       newOption(),
	}
	app.baseCtx, app.baseCtxCancel = context.WithCancel(context.Background())

	// 初始化选项
	for _, o := range opts {
		o(app.opt)
	}

	// 处理选项
	if app.opt.EnableDaemon {
		app.enableDaemon()
	}

	app.config = config.NewConfig(appName, app.opt.ConfigOpts...)
	app.ILogger = logger.NewLogger(appName, app.config)

	app.Debug("app初始化")
	app.handler(BeforeInitializeHandler)

	// 初始化组件
	app.component = component.NewComponent(app)

	// 初始化服务
	app.opt.CheckCustomEnableServices(app.component)
	for serviceType, enable := range app.opt.Services {
		if enable {
			app.services[serviceType] = service.MakeService(app, serviceType)
		}
	}

	app.handler(AfterInitializeHandler)
	app.Debug("app初始化完毕")

	return app
}

func (app *appCli) run() {
	app.Debug("启动app")
	app.handler(BeforeStartHandler)

	// 启动服务
	app.startService()

	go app.freeMemory()

	app.handler(AfterStartHandler)
	app.Info("app已启动")

	signal.Notify(app.interrupt, os.Kill, os.Interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-app.interrupt
	app.exit()
}

func (app *appCli) startService() {
	app.Debug("启动服务")
	for serviceType, s := range app.services {
		if err := s.Start(); err != nil {
			app.Fatal("服务启动失败", zap.String("serviceType", string(serviceType)), zap.Error(err))
		}
	}
}

func (app *appCli) closeService() {
	app.Debug("关闭服务")
	for serviceType, s := range app.services {
		if err := s.Close(); err != nil {
			app.Error("服务关闭失败", zap.String("serviceType", string(serviceType)), zap.Error(err))
		}
	}
}

func (app *appCli) enableDaemon() {
	if len(os.Args) < 2 {
		return
	}

	switch os.Args[1] {
	case "install":
	case "remove":
	case "start":
	case "stop":
	case "status":
	default:
		return
	}

	d, err := daemon.New(app.name, app.name, daemon.SystemDaemon)
	if err != nil {
		logger.Log.Fatal("守护进程模块创建失败", zap.Error(err))
	}

	var out string
	switch os.Args[1] {
	case "install":
		out, err = d.Install(os.Args[2:]...)
	case "remove":
		out, err = d.Remove()
	case "start":
		out, err = d.Start()
	case "stop":
		out, err = d.Stop()
	case "status":
		out, err = d.Status()
	}

	if err != nil {
		logger.Log.Error(out, zap.Error(err))
		os.Exit(1)
	}

	logger.Log.Info(out)
	os.Exit(0)
}

func (app *appCli) exit() {
	app.Debug("app准备退出")

	// app退出前
	app.handler(BeforeExitHandler)

	// 关闭基础上下文
	app.baseCtxCancel()
	// 关闭服务
	app.closeService()
	// 释放组件资源
	app.closeComponentResource()

	// app退出后
	app.handler(AfterExitHandler)
	app.Warn("app已退出")
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
	app.interrupt <- syscall.SIGTERM
}

func (app *appCli) GetConfig() core.IConfig {
	return app.config
}

func (app *appCli) BaseContext() context.Context {
	return app.baseCtx
}

func (app *appCli) GetLogger() core.ILogger {
	return app.ILogger
}

func (app *appCli) freeMemory() {
	interval := app.config.Config().FreeMemoryInterval
	if interval <= 0 {
		return
	}

	t := time.NewTicker(time.Duration(interval) * time.Millisecond)
	for {
		select {
		case <-app.baseCtx.Done():
			t.Stop()
			return
		case <-t.C:
			debug.FreeOSMemory()
		}
	}
}

func (app *appCli) handler(t HandlerType) {
	for _, h := range app.opt.Handlers[t] {
		h(app, t)
	}
}
