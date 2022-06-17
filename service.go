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
	err := app.opt.CheckServices(app)
	if err != nil {
		app.Fatal("服务检查失败", zap.Error(err))
	}

	for _, serviceType := range app.opt.Services {
		app.services[serviceType] = service.MakeService(app, serviceType)
	}
}

func (app *appCli) startService() {
	app.Debug("启动服务")
	for _, serviceType := range app.opt.Services {
		s, ok := app.services[serviceType]
		if !ok {
			app.Fatal("服务查找失败", zap.String("serviceType", string(serviceType)))
		}

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
	for _, serviceType := range app.opt.Services {
		s, ok := app.services[serviceType]
		if !ok {
			app.Fatal("服务查找失败", zap.String("serviceType", string(serviceType)))
		}

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
		logger.Log.Fatal("注入失败, 未启用服务", zap.String("serviceType", string(serviceType)))
	}

	s.Inject(a...)
}
