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
	"github.com/zly-app/zapp/pkg/depender"
	"github.com/zly-app/zapp/plugin"
)

// 初始化插件
func (app *appCli) makePlugin() {
	err := app.opt.CheckPlugins(app)
	if err != nil {
		app.Fatal("插件检查失败", zap.Error(err))
	}

	for _, pluginType := range app.opt.Plugins {
		app.plugins[pluginType] = plugin.MakePlugin(app, pluginType)
	}
}

func (app *appCli) startPlugin() {
	app.Debug("启动插件")
	type Item struct {
		Name      string
		DependsOn []string
		Close     func()
	}
	items := make([]depender.Item, len(app.opt.Plugins))
	for i, pluginType := range app.opt.Plugins {
		p, ok := app.plugins[pluginType]
		if !ok {
			app.Fatal("插件查找失败", zap.String("pluginType", string(pluginType)))
		}
		var dps []string = nil
		if dp, ok := p.(core.Depender); ok {
			dps = dp.DependsOn()
		}
		items[i] = depender.NewItem(string(pluginType), dps, func() error {
			return p.Start()
		}, func() {
			err := p.Close()
			if err != nil {
				app.Error("插件关闭失败", zap.String("pluginType", string(pluginType)), zap.Error(err))
			}
		})
	}
	dep := depender.NewDepender(items)
	err := dep.Start()
	if err != nil {
		app.Fatal("插件启动失败", zap.Error(err))
	}
	app.pluginsDepender = dep
}

func (app *appCli) closePlugin() {
	app.Debug("关闭插件")
	if app.pluginsDepender != nil {
		app.pluginsDepender.Close()
		return
	}

	// 可能没有调用 app.Run 这里需要主动遍历关闭
	for _, pluginType := range app.opt.Plugins {
		p, ok := app.plugins[pluginType]
		if !ok {
			app.Fatal("插件查找失败", zap.String("pluginType", string(pluginType)))
		}

		if err := p.Close(); err != nil {
			app.Error("插件关闭失败", zap.String("pluginType", string(pluginType)), zap.Error(err))
		}
	}
}

func (app *appCli) GetPlugin(pluginType core.PluginType) (core.IPlugin, bool) {
	p, ok := app.plugins[pluginType]
	return p, ok
}

func (app *appCli) InjectPlugin(pluginType core.PluginType, a ...interface{}) {
	p, ok := app.GetPlugin(pluginType)
	if !ok {
		if app.opt.IgnoreInjectOfDisablePlugin {
			return
		}
		logger.Log.Fatal("注入失败, 未启用插件", zap.String("pluginType", string(pluginType)))
	}

	p.Inject(a...)
}
