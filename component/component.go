package component

import (
	"github.com/zly-app/zapp/component/gpool"
	"github.com/zly-app/zapp/component/msgbus"
	"github.com/zly-app/zapp/core"
)

// ComponentCli 是默认的组件实现：把协程池与消息总线装配成一个组件集合。
//
// 它自身不持有可变状态，只做「把已有能力暴露成 core.IComponent」这一件事。
type ComponentCli struct {
	app    core.IApp
	config *core.Config
	core.ILogger

	core.IGPools
	core.IMsgbus
}

func (c *ComponentCli) App() core.IApp       { return c.app }
func (c *ComponentCli) Config() *core.Config { return c.config }

func (c *ComponentCli) Close() {}

// NewComponent 构建默认组件集合
//
// 这里不写入任何包级全局变量：组件实例由 Container 持有、经 App 暴露。
func NewComponent(app core.IApp) core.IComponent {
	return &ComponentCli{
		app:     app,
		config:  app.GetConfig().Config(),
		ILogger: app.GetLogger(),

		IGPools: gpool.GetCreator(),
		IMsgbus: msgbus.GetMsgbus(),
	}
}
