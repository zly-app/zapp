package config

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

// 选择provider
func WithWatchProvider(name string) core.ConfigWatchOption {
	return func(w interface{}) {
		p := GetConfigWatchProvider(name)
		if p == nil {
			logger.Log.Fatal("配置观察提供者不存在", zap.String("name", name))
		}

		watcher := w.(*watchKeyObject)
		watcher.p = p
	}
}
