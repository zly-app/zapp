package config

import (
	"github.com/zly-app/zapp/core"
)

// 观察key, 失败会fatal
func (c *configCli) WatchKey(groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyObject {
	return WatchKey(groupName, keyName, opts...)
}
