package plugin

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

// 插件建造者
type pluginCreator func(app core.IApp) core.IPlugin

func (h pluginCreator) Create(app core.IApp) core.IPlugin {
	return h(app)
}

// 建造者列表
var creators = make(map[core.PluginType]core.IPluginCreator)

// 注册插件建造者
func RegisterCreator(pluginType core.PluginType, creator core.IPluginCreator) {
	if _, ok := creators[pluginType]; ok {
		logger.Log.Fatal("重复注册建造者", zap.String("pluginType", string(pluginType)))
	}
	creators[pluginType] = creator
}

// 注册插件建造者函数
func RegisterCreatorFunc(pluginType core.PluginType, creatorFunc func(app core.IApp) core.IPlugin) {
	RegisterCreator(pluginType, pluginCreator(creatorFunc))
}

// 构建插件
func MakePlugin(app core.IApp, pluginType core.PluginType) core.IPlugin {
	if creator, ok := creators[pluginType]; ok {
		return creator.Create(app)
	}
	app.Fatal("使用了未注册建造者的插件", zap.String("pluginType", string(pluginType)))
	return nil
}
