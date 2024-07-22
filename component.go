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
	app.Info("构建组件")
	app.handler(BeforeMakeComponent)

	app.component = component.NewComponent(app)
	for _, fn := range app.opt.CustomComponentFn {
		app.component = fn(app, app.component)
	}
	component.ResetComponent(app.component)

	app.handler(AfterMakeComponent)
}

// 释放组件资源
func (app *appCli) releaseComponentResource() {
	app.Info("释放组件资源")
	app.handler(BeforeCloseComponent)
	app.component.Close()
	app.handler(AfterCloseComponent)
}

func (app *appCli) GetComponent() core.IComponent {
	return app.component
}
