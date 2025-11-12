package config

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/log"
)

type StructType string

const (
	Json StructType = "json"
	Yaml StructType = "yaml"
)

type watchOptions struct {
	Provider   core.IConfigWatchProvider
	StructType StructType
}

func newWatchOptions(opts []core.ConfigWatchOption) *watchOptions {
	o := &watchOptions{}
	for _, fn := range opts {
		fn(o)
	}

	if o.StructType == "" {
		o.StructType = Json
	}
	o.check()
	return o
}

// 检查状态
func (o *watchOptions) check() {
	if o.Provider == nil {
		o.Provider = GetDefaultConfigWatchProvider()
	}
	if o.Provider == nil {
		log.Log.Fatal("默认配置观察提供者不存在")
	}
}

func getWatchOptions(a interface{}) *watchOptions {
	opts, ok := a.(*watchOptions)
	if !ok {
		log.Log.Fatal("无法转换为*watchOptions", zap.Any("a", a))
	}
	return opts
}

// 选择provider
func WithWatchProvider(name string) core.ConfigWatchOption {
	return func(a interface{}) {
		p := GetConfigWatchProvider(name)
		if p == nil {
			log.Log.Fatal("配置观察提供者不存在", zap.String("name", name))
		}

		opts := getWatchOptions(a)
		opts.Provider = p
	}
}

// 设置观察结构化类型
func WithWatchStructType(t StructType) core.ConfigWatchOption {
	return func(a interface{}) {
		opts := getWatchOptions(a)
		opts.StructType = t
	}
}

// 设置观察结构化类型为json
func WithWatchStructJson() core.ConfigWatchOption {
	return func(a interface{}) {
		opts := getWatchOptions(a)
		opts.StructType = Json
	}
}

// 设置观察结构化类型为Yaml
func WithWatchStructYaml() core.ConfigWatchOption {
	return func(a interface{}) {
		opts := getWatchOptions(a)
		opts.StructType = Yaml
	}
}
