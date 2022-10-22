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

// 当前服务类型
var nowPluginType = DefaultPluginType

// 设置插件类型, 这个函数应该在 zapp.NewApp 之前调用
func SetPluginType(t core.PluginType) {
	nowPluginType = t
}

// 启用插件, 用于设置配置观察的提供者
func WithPlugin(setDefaultProvider ...bool) zapp.Option {
	plugin.RegisterCreatorFunc(nowPluginType, func(app core.IApp) core.IPlugin {
		p := NewApolloProvider(app)
		config.RegistryConfigWatchProvider(ProviderName, p) // 注册提供者
		if len(setDefaultProvider) > 0 && setDefaultProvider[0] {
			config.SetDefaultConfigWatchProvider(p) // 设为默认
		}
		return p
	})
	return zapp.WithPlugin(nowPluginType)
}
