package apollo_provider

import (
	"github.com/zly-app/zapp"
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/plugin"
)

// 提供者名
const ProviderName = "apollo"

// 默认插件类型
const DefaultPluginType core.PluginType = "apollo_provider"

var _setDefaultProvider bool

func init() {
	plugin.RegisterCreatorFunc(DefaultPluginType, func(app core.IApp) core.IPlugin {
		p := NewApolloProvider(app)
		config.RegistryConfigWatchProvider(ProviderName, p) // 注册提供者
		if _setDefaultProvider {
			config.SetDefaultConfigWatchProvider(p) // 设为默认
		}
		return p
	})
}

// 启用插件, 用于设置配置观察的提供者
func WithPlugin(setDefaultProvider ...bool) zapp.Option {
	if len(setDefaultProvider) > 0 && setDefaultProvider[0] {
		_setDefaultProvider = true // 任何一次将其设为默认
	}
	return zapp.WithPlugin(DefaultPluginType)
}
