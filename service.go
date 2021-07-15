/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/20
   Description :
-------------------------------------------------
*/

package zapp

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/service"
)

// 构建服务
func (app *appCli) makeService() {
	app.opt.CheckCustomEnableServices(app)
	for serviceType, enable := range app.opt.Services {
		if enable {
			app.services[serviceType] = service.MakeService(app, serviceType)
		}
	}
}

func (app *appCli) startService() {
	app.Debug("启动服务")
	for serviceType, s := range app.services {
		err := service.WaitRun(app, &service.WaitRunOption{
			ServiceType:        serviceType,
			ExitOnErrOfObserve: app.opt.ExitOnErrOfObserveServiceUnstable,
			RunServiceFn:       s.Start,
		})
		if err != nil {
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

func (app *appCli) GetService(serviceType core.ServiceType) (core.IService, bool) {
	s, ok := app.services[serviceType]
	return s, ok
}

func (app *appCli) InjectService(serviceType core.ServiceType, a ...interface{}) {
	s, ok := app.GetService(serviceType)
	if !ok {
		if app.opt.IgnoreInjectOfDisableService {
			return
		}
		logger.Log.Fatal("未启用api服务")
	}

	s.Inject(a...)
}
