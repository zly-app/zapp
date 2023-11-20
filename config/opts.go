/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package config

import (
	"github.com/spf13/viper"

	"github.com/zly-app/zapp/core"
)

type Option func(o *Options)

type Options struct {
	vi           *viper.Viper  // viper
	conf         *core.Config  // 配置结构
	files        []string      // 配置文件
	apolloConfig *ApolloConfig // apollo配置结构
	disableFlag  bool          // 不启用flag
}

func newOptions() *Options {
	return &Options{}
}

// 设置viper, 优先级低于从命令行指定配置文件
func WithViper(vi *viper.Viper) Option {
	return func(o *Options) {
		o.vi = vi
	}
}

// 主动设置配置结构, 优先级低于WithViper
func WithConfig(conf *core.Config) Option {
	return func(o *Options) {
		o.conf = conf
	}
}

// 主动设置配置文件, 优先级低于WithConfig
func WithFiles(files ...string) Option {
	return func(o *Options) {
		o.files = files
	}
}

// 从apollo加载配置, 必须是规范的设置
func WithApollo(conf *ApolloConfig) Option {
	return func(o *Options) {
		o.apolloConfig = conf
	}
}

// 不启用 flag
func WithoutFlag() Option {
	return func(o *Options) {
		o.disableFlag = true
	}
}
