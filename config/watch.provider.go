package config

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

// 用于承载配置观察提供者
var configWatchProviders = make(map[string]core.IConfigWatchProvider)

// 获取默认配置观察提供者, 可能返回nil
func GetDefaultConfigWatchProvider() core.IConfigWatchProvider {
	return configWatchProviders["default"]
}

// 设置默认配置观察提供者
func SetDefaultConfigWatchProvider(p core.IConfigWatchProvider) {
	configWatchProviders["default"] = p
}

// 获取配置观察者, 不存在时返回nil
func GetConfigWatchProvider(name string) core.IConfigWatchProvider {
	return configWatchProviders[name]
}

// 注册配置观察提供者
func RegistryConfigWatchProvider(name string, p core.IConfigWatchProvider) {
	if name == "default" {
		logger.Log.Fatal("配置提供者名不能为default")
	}
	if _, ok := configWatchProviders[name]; ok {
		logger.Log.Fatal("配置提供者已存在", zap.String("name", name))
	}
	configWatchProviders[name] = p
}
