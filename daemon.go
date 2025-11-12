package zapp

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/kardianos/service"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/log"
)

type daemonService struct {
	app *appCli
}

func (m *daemonService) Start(s service.Service) error {
	go m.app.run()
	return nil
}

func (m *daemonService) Stop(s service.Service) error {
	m.app.Exit()
	return nil
}

func (app *appCli) setDaemonService() {
	svcConfig := &service.Config{
		Name:      app.name,
		Arguments: os.Args[1:],
	}
	prg := &daemonService{app: app}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	app.daemonService = s
}

func (app *appCli) enableDaemon() {
	if !app.opt.EnableDaemon || len(os.Args) < 2 {
		app.setDaemonService()
		return
	}

	flag.String("install", "", "安装服务, string 是运行时传递给 app 的参数, 请不要使用相对路径")
	flag.Bool("start", false, "启动服务")
	flag.Bool("stop", false, "停止服务")
	flag.Bool("restart", false, "重启服务")
	flag.Bool("status", false, "获取状态")
	flag.Bool("uninstall", false, "移除服务")

	svcConfig := &service.Config{
		Name: app.name,
	}

	switch os.Args[1] {
	case "install", "-install":
		svcConfig.Arguments = os.Args[2:]
	case "remove", "start", "stop", "restart", "status", "uninstall":
	case "-remove", "-start", "-stop", "-restart", "-status", "-uninstall":
	default:
		app.setDaemonService()
		return
	}

	prg := &daemonService{app: app}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "install", "-install":
		err = s.Install()
	case "start", "-start":
		err = s.Start()
	case "stop", "-stop":
		err = s.Stop()
	case "restart", "-restart":
		err = s.Restart()
	case "uninstall", "-uninstall":
		err = s.Uninstall()
	case "status", "-status":
		status, err := s.Status()
		if err != nil {
			log.Log.Error(zap.Error(err))
			os.Exit(1)
		}

		switch status {
		case service.StatusRunning:
			log.Log.Info(fmt.Sprintf("service %s is Running", app.name))
		case service.StatusStopped:
			log.Log.Info(fmt.Sprintf("service %s is Stopped", app.name))
		default:
			log.Log.Info(fmt.Sprintf("service %s status is Unknown", app.name))
		}
		os.Exit(0)
	}

	if err != nil {
		log.Log.Error(zap.Error(err))
		os.Exit(1)
	}

	log.Log.Info("ok")
	os.Exit(0)
}

func Run() {
	log.Info("Service start...")
	// 在这里编写你的服务逻辑
	for {
		log.Info("Service is running...")
		time.Sleep(1 * time.Second)
	}
}
