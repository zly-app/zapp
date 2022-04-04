package config

import (
	"github.com/zly-app/zapp/core"
)

// 观察key
func (c *configCli) WatchKey(groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyObject {
	panic("implement me")
}
