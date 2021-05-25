/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/20
   Description :
-------------------------------------------------
*/

package component

import (
	"github.com/zly-app/zapp/component/gpool"
	"github.com/zly-app/zapp/component/msgbus"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

var defaultComponent core.IComponent

type ComponentCli struct {
	app    core.IApp
	config *core.Config
	core.ILogger

	core.IGPool
	core.IMsgbus
}

func (c *ComponentCli) App() core.IApp       { return c.app }
func (c *ComponentCli) Config() *core.Config { return c.config }

func (c *ComponentCli) Close() {
	c.IGPool.Close()
	c.IMsgbus.Close()
}

func NewComponent(app core.IApp) core.IComponent {
	var c core.IComponent = &ComponentCli{
		app:     app,
		config:  app.GetConfig().Config(),
		ILogger: app.GetLogger(),

		IGPool:  gpool.NewGPoolManager(),
		IMsgbus: msgbus.NewMsgbus(),
	}
	defaultComponent = c
	return c
}

// 获取component
func GetComponent() core.IComponent {
	if defaultComponent == nil {
		logger.Log.Panic("Component is uninitialized")
	}
	return defaultComponent
}

// 重置component
func ResetComponent(component core.IComponent) {
	defaultComponent = component
}
