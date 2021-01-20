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
)

func (app *appCli) GetService(serviceType core.ServiceType) (core.IService, bool) {
	s, ok := app.services[serviceType]
	return s, ok
}
