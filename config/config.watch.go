package config

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

// 用于承载配置观察提供者
var configWatchProviders = make(map[string]core.ConfigWatchProvider)

// 获取默认配置观察提供者, 可能返回nil
func GetDefaultConfigWatchProvider() core.ConfigWatchProvider {
	return configWatchProviders["default"]
}

// 设置默认配置观察提供者
func SetDefaultConfigWatchProvider(p core.ConfigWatchProvider) {
	configWatchProviders["default"] = p
}

// 添加配置观察提供者, 第一个被添加的提供者会作为默认提供者
func AddConfigWatchProvider(name string, p core.ConfigWatchProvider) {
	if name == "default" {
		logger.Log.Fatal("配置提供者名不能为default")
	}
	if _, ok := configWatchProviders[name]; ok {
		logger.Log.Fatal("配置提供者已存在", zap.String("name", name))
	}
	configWatchProviders[name] = p

	// 设置默认配置提供者
	if GetDefaultConfigWatchProvider() == nil {
		SetDefaultConfigWatchProvider(p)
	}
}
