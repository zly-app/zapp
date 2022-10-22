package zapp

import (
	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
)

// 观察key, 失败会fatal, 支持在定义变量时初始化
func WatchConfigKey(groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyObject {
	return config.WatchKey(groupName, keyName, opts...)
}

// 观察json配置数据, 失败会fatal, 支持在定义变量时初始化
func WatchConfigJson[T any](groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyStruct[T] {
	return config.WatchJson[T](groupName, keyName, opts...)
}

// 观察yaml配置数据, 失败会fatal, 支持在定义变量时初始化
func WatchConfigYaml[T any](groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyStruct[T] {
	return config.WatchYaml[T](groupName, keyName, opts...)
}
