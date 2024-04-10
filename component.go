/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/20
   Description :
-------------------------------------------------
*/

package zapp

import (
	"github.com/zly-app/zapp/component"
	"github.com/zly-app/zapp/core"
)

// 构建组件
func (app *appCli) makeComponent() {
	app.component = component.NewComponent(app)
	for _, fn := range app.opt.CustomComponentFn {
		app.component = fn(app, app.component)
	}
	component.ResetComponent(app.component)
}

// 释放组件资源
func (app *appCli) releaseComponentResource() {
	app.Debug("释放组件资源")
	app.component.Close()
}

func (app *appCli) GetComponent() core.IComponent {
	return app.component
}
