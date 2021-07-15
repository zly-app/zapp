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
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
	"github.com/zly-app/zapp/pkg/zlog"
)

var Conf core.IConfig

type configCli struct {
	vi     *viper.Viper
	c      *core.Config
	flags  map[string]struct{}
	labels map[string]string
}

func newConfig(appName string) *core.Config {
	conf := &core.Config{
		Frame: core.FrameConfig{
			Debug:              true,
			FreeMemoryInterval: consts.DefaultFreeMemoryInterval,
			Labels:             make(map[string]string),
			Log:                zlog.DefaultConfig,
		},
	}
	conf.Frame.Log.Name = appName
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

	var rawVi *viper.Viper
	var err error
	if *confText != "" { // 命令行
		files := strings.Split(*confText, ",")
		rawVi, err = makeViperFromFile(files, false)
		if err != nil {
			logger.Log.Fatal("从命令指定文件构建viper失败", zap.Strings("files", files), zap.Error(err))
		}
	} else if opt.vi != nil { // WithViper
		rawVi = opt.vi
	} else if opt.conf != nil { // WithConfig
		rawVi, err = makeViperFromStruct(opt.conf)
		if err != nil {
			logger.Log.Fatal("从配置结构构建viper失败", zap.Any("config", opt.conf), zap.Error(err))
		}
	} else if len(opt.files) > 0 { // WithFiles
		rawVi, err = makeViperFromFile(opt.files, false)
		if err != nil {
			logger.Log.Fatal("从用户指定文件构建viper失败", zap.Strings("files", opt.files), zap.Error(err))
		}
	} else if opt.apolloConfig != nil { // WithApollo
		rawVi, err = makeViperFromApollo(opt.apolloConfig)
		if err != nil {
			logger.Log.Fatal("从apollo构建viper失败", zap.Error(err))
		}
	} else { // 默认
		files := strings.Split(consts.DefaultConfigFiles, ",")
		logger.Log.Debug("使用默认配置文件", zap.Strings("files", files))
		rawVi, err = makeViperFromFile(files, true)
		if err != nil {
			logger.Log.Fatal("从默认配置文件构建viper失败", zap.Strings("files", files), zap.Error(err))
		}
	}

	vi := viper.New()
	vi.MergeConfigMap(rawVi.AllSettings())

	// 如果从viper中发现了apollo配置
	if vi.IsSet(consts.ApolloConfigKey) {
		apolloConf, err := makeApolloConfigFromViper(vi)
		if err != nil {
			logger.Log.Fatal("解析apollo配置失败", zap.Error(err))
		}
		rawVi, err = makeViperFromApollo(apolloConf)
		if err != nil {
			logger.Log.Fatal("从apollo构建viper失败", zap.Error(err))
		}
		if err = vi.MergeConfigMap(rawVi.AllSettings()); err != nil {
			logger.Log.Fatal("合并apollo配置失败", zap.Error(err))
		}
	}

	c := &configCli{
		vi: vi,
		c:  newConfig(appName),
	}
	// 解析配置
	if err = vi.Unmarshal(c.c); err != nil {
		logger.Log.Fatal("配置解析失败", zap.Error(err))
	}

	c.checkDefaultConfig(c.c)

	if *testFlag {
		fmt.Println("配置文件测试成功")
		os.Exit(0)
	}

	c.makeFlags()
	c.makeLabels()

	Conf = c
	return c
}

func (c *configCli) makeFlags() {
	c.flags = make(map[string]struct{}, len(c.c.Frame.Flags))
	for _, flag := range c.c.Frame.Flags {
		c.flags[strings.ToLower(flag)] = struct{}{}
	}

	flags := make([]string, 0, len(c.flags))
	for flag := range c.flags {
		flags = append(flags, flag)
	}
	c.c.Frame.Flags = flags
}

func (c *configCli) makeLabels() {
	c.labels = make(map[string]string, len(c.c.Frame.Labels))
	for k, v := range c.c.Frame.Labels {
		c.labels[strings.ToLower(k)] = v
	}
	c.c.Frame.Labels = c.labels
}

// 从文件构建viper
func makeViperFromFile(files []string, isDefault bool) (*viper.Viper, error) {
	vi := viper.New()
	for _, file := range files {
		_, err := os.Stat(file)
		if err != nil {
			if isDefault && os.IsNotExist(err) { // 如果默认配置文件不存在则忽略
				logger.Log.Warn("默认配置文件不存在", zap.String("file", file))
				continue
			}
			return nil, fmt.Errorf("读取配置文件'%s'信息失败: %s", file, err)
		}

		vi.SetConfigFile(file)
		if err = vi.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("合并配置文件'%s'失败: %s", file, err)
		}
	}
	return vi, nil
}

// 从结构体构建viper
func makeViperFromStruct(a interface{}) (*viper.Viper, error) {
	bs, err := json.Marshal(a)
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
	conf.Frame.WaitServiceRunTime = utils.Ternary.Or(conf.Frame.WaitServiceRunTime, consts.DefaultWaitServiceRunTime).(int)
	conf.Frame.ServiceUnstableObserveTime = utils.Ternary.Or(conf.Frame.ServiceUnstableObserveTime, consts.DefaultServiceUnstableObserveTime).(int)
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
	if err := c.vi.UnmarshalKey(key, outPtr); err != nil {
		return fmt.Errorf("无法解析key<%s>配置: %s", key, err)
	}
	return nil
}

func (c *configCli) ParseComponentConfig(componentType core.ComponentType, componentName string, outPtr interface{}) error {
	componentName = utils.Ternary.Or(componentName, consts.DefaultComponentName).(string)
	key := "components." + string(componentType) + "." + componentName
	if !c.vi.IsSet(key) {
		return fmt.Errorf("组件配置<%s.%s>不存在", componentType, componentName)
	}
	if err := c.vi.UnmarshalKey(key, outPtr); err != nil {
		return fmt.Errorf("无法解析<%s.%s>组件配置: %s", componentType, componentName, err)
	}
	return nil
}

func (c *configCli) ParsePluginConfig(pluginType core.PluginType, outPtr interface{}) error {
	key := "plugins." + string(pluginType)
	if !c.vi.IsSet(key) {
		return fmt.Errorf("插件配置<%s>不存在", pluginType)
	}
	if err := c.vi.UnmarshalKey(key, outPtr); err != nil {
		return fmt.Errorf("无法解析<%s>插件配置: %s", pluginType, err)
	}
	return nil
}

func (c *configCli) ParseServiceConfig(serviceType core.ServiceType, outPtr interface{}) error {
	key := "services." + string(serviceType)
	if !c.vi.IsSet(key) {
		return fmt.Errorf("服务配置<%s>不存在", serviceType)
	}
	if err := c.vi.UnmarshalKey(key, outPtr); err != nil {
		return fmt.Errorf("无法解析<%s>服务配置: %s", serviceType, err)
	}
	return nil
}

func (c *configCli) GetLabel(name string) string {
	return c.labels[strings.ToLower(name)]
}

func (c *configCli) GetLabels() map[string]string {
	return c.labels
}

func (c *configCli) HasFlag(flag string) bool {
	_, ok := c.flags[strings.ToLower(flag)]
	return ok
}

func (c *configCli) GetFlags() []string {
	return c.c.Frame.Flags
}
