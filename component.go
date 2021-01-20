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

// 关闭组件内加载的资源
func (app *appCli) closeComponentResource() {
	app.Debug("释放组件加载的资源")
	app.component.Close()
}

func (app *appCli) GetComponent() core.IComponent {
	return app.component
}
