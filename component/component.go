/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/20
   Description :
-------------------------------------------------
*/

package component

import (
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

var defaultComponent core.IComponent

type ComponentCli struct {
	app    core.IApp
	config *core.Config
	core.ILogger
}

func (c *ComponentCli) App() core.IApp       { return c.app }
func (c *ComponentCli) Config() *core.Config { return c.config }

func (c *ComponentCli) Close() {
}

func NewComponent(app core.IApp) core.IComponent {
	var c core.IComponent = &ComponentCli{
		app:     app,
		config:  app.GetConfig().Config(),
		ILogger: app.GetLogger(),
	}
	defaultComponent = c
	return c
}

// 获取全局component
func GlobalComponent() core.IComponent {
	if defaultComponent == nil {
		logger.Log.Panic("GlobalComponent is uninitialized")
	}
	return defaultComponent
}

// 重置全局component
func ResetGlobalComponent(app core.IApp) {
	defaultComponent = app.GetComponent()
}
