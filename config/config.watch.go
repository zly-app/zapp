package config

import (
	"github.com/zly-app/zapp/core"
)

// 观察key
func (c *configCli) WatchKey(groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyObject {
	w := newWatchKeyObject(groupName, keyName, opts...)
	return w
}
