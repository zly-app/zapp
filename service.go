/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/20
   Description :
-------------------------------------------------
*/

package zapp

import (
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

func (app *appCli) GetService(serviceType core.ServiceType) (core.IService, bool) {
	s, ok := app.services[serviceType]
	return s, ok
}

func (app *appCli) InjectService(serviceType core.ServiceType, a interface{}) {
	s, ok := app.GetService(serviceType)
	if !ok {
		if app.opt.IgnoreInjectOfDisableServer {
			return
		}
		logger.Log.Fatal("未启用api服务")
	}

	s.Inject(a)
}
