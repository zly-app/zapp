/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/20
   Description :
-------------------------------------------------
*/

package config

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"github.com/zlyuancn/zutils"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

var Config core.IConfig

type configCli struct {
	vi     *viper.Viper
	c      *core.Config
	labels map[string]interface{}
}

func newConfig() *core.Config {
	conf := &core.Config{
		Debug:              true,
		FreeMemoryInterval: consts.DefaultFreeMemoryInterval,
		Labels:             make(map[string]interface{}),
	}
	return conf
}

// 解析配置
//
// 配置来源优先级 命令行 > WithViper > WithConfig > WithFiles(Apollo分片优先级最高) > WithApollo > 默认配置文件
// 注意: 多个配置文件如果存在同配置分片会智能合并, 同分片中完全相同的配置节点以最后的文件为准, 从apollo拉取的配置会覆盖相同的文件配置节点
func NewConfig(appName string, opts ...Option) core.IConfig {
	opt := newOptions()
	for _, o := range opts {
		o(opt)
	}

	confText := flag.String("c", "", "配置文件,多个文件用逗号隔开,同名配置分片会完全覆盖之前的分片")
	testFlag := flag.Bool("t", false, "测试配置文件")
	flag.Parse()

	var vi *viper.Viper
	var err error
	if *confText != "" { // 命令行
		files := strings.Split(*confText, ",")
		vi, err = makeViperFromFile(files)
		if err != nil {
			logger.Log.Fatal("从命令指定文件构建viper失败", zap.Strings("files", files), zap.Error(err))
		}
	} else if opt.vi != nil { // WithViper
		vi = opt.vi
	} else if opt.conf != nil { // WithConfig
		vi, err = makeViperFromStruct(opt.conf)
		if err != nil {
			logger.Log.Fatal("从配置结构构建viper失败", zap.Any("config", opt.conf), zap.Error(err))
		}
	} else if len(opt.files) > 0 { // WithFiles
		vi, err = makeViperFromFile(opt.files)
		if err != nil {
			logger.Log.Fatal("从用户指定文件构建viper失败", zap.Strings("files", opt.files), zap.Error(err))
		}
	} else if opt.apolloConfig != nil { // WithApollo
		vi, err = makeViperFromApollo(opt.apolloConfig)
		if err != nil {
			logger.Log.Fatal("从apollo构建viper失败", zap.Any("apolloConfig", opt.apolloConfig), zap.Error(err))
		}
	} else { // 默认
		files := strings.Split(consts.DefaultConfigFiles, ",")
		logger.Log.Debug("使用默认配置文件", zap.Strings("files", files))
		vi, err = makeViperFromFile(files)
		if err != nil {
			logger.Log.Fatal("从默认配置文件构建viper失败", zap.Strings("files", files), zap.Error(err))
		}
	}

	// 如果从viper中发现了apollo配置
	if vi.IsSet(consts.ApolloConfigKey) {
		apolloConf, err := makeApolloConfigFromViper(vi)
		if err != nil {
			logger.Log.Fatal("解析apollo配置失败", zap.Error(err))
		}
		newVi, err := makeViperFromApollo(apolloConf)
		if err != nil {
			logger.Log.Fatal("从apollo构建viper失败", zap.Any("apolloConfig", apolloConf), zap.Error(err))
		}
		if err = vi.MergeConfigMap(newVi.AllSettings()); err != nil {
			logger.Log.Fatal("合并apollo配置失败", zap.Error(err))
		}
	}

	c := &configCli{
		vi: vi,
		c:  newConfig(),
	}
	if err = vi.UnmarshalKey(consts.FrameConfigKey, c.c); err != nil {
		logger.Log.Fatal("配置解析失败", zap.Error(err))
	}

	c.checkDefaultConfig(c.c)

	if *testFlag {
		fmt.Println("配置文件测试成功")
		os.Exit(0)
	}

	c.makeTags()
	Config = c
	return c
}

func (c *configCli) makeTags() {
	c.labels = make(map[string]interface{}, len(c.c.Labels))
	for k, v := range c.c.Labels {
		c.labels[strings.ToLower(k)] = v
	}
}

// 从文件构建viper
func makeViperFromFile(files []string) (*viper.Viper, error) {
	vi := viper.New()
	for _, file := range files {
		vp := viper.New()
		vp.SetConfigFile(file)
		if err := vp.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("配置文件'%s'加载失败: %s", file, err)
		}
		if err := vi.MergeConfigMap(vp.AllSettings()); err != nil {
			return nil, fmt.Errorf("合并配置文件'%s'失败: %s", file, err)
		}
	}
	return vi, nil
}

// 从结构体构建viper
func makeViperFromStruct(a interface{}) (*viper.Viper, error) {
	bs, err := jsoniter.Marshal(a)
	if err != nil {
		return nil, fmt.Errorf("编码失败: %s", err)
	}

	vi := viper.New()
	vi.SetConfigType("json")
	err = vi.ReadConfig(bytes.NewReader(bs))
	if err != nil {
		return nil, fmt.Errorf("数据解析失败: %s", err)
	}
	return vi, nil
}

func (c *configCli) checkDefaultConfig(conf *core.Config) {
	conf.WaitServiceRunTime = zutils.Ternary.Or(conf.WaitServiceRunTime, consts.DefaultWaitServiceRunTime).(int)
	conf.ServiceUnstableObserveTime = zutils.Ternary.Or(conf.ServiceUnstableObserveTime, consts.DefaultServiceUnstableObserveTime).(int)
}

func (c *configCli) Config() *core.Config {
	return c.c
}

func (c *configCli) GetViper() *viper.Viper {
	return c.vi
}

func (c *configCli) Parse(key string, outPtr interface{}) error {
	if !c.vi.IsSet(key) {
		return fmt.Errorf("key<%s>不存在", key)
	}
	return c.vi.UnmarshalKey(key, outPtr)
}

func (c *configCli) ParseServiceConfig(serviceType string, outPtr interface{}) error {
	key := "services." + serviceType
	if !c.vi.IsSet(key) {
		return fmt.Errorf("服务配置<%s>不存在", serviceType)
	}
	return c.vi.UnmarshalKey(key, outPtr)
}

func (c *configCli) ParseComponentConfig(componentType, componentName string, outPtr interface{}) error {
	key := "components." + componentType + "." + componentName
	if !c.vi.IsSet(key) {
		return fmt.Errorf("组件配置<%s.%s>不存在", componentType, componentName)
	}
	return c.vi.UnmarshalKey(key, outPtr)
}

func (c *configCli) GetLabel(name string) interface{} {
	return c.labels[strings.ToLower(name)]
}
