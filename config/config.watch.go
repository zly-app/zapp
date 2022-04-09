package config

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

// 观察key
func (c *configCli) WatchKey(groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyObject {
	w, err := newWatchKeyObject(groupName, keyName, opts...)
	if err != nil {
		logger.Log.Fatal("观察key失败",
			zap.String("groupName", groupName),
			zap.String("keyName", keyName),
			zap.Error(err),
		)
	}
	return w
}
