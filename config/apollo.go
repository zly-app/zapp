/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package config

import (
	"encoding/base64"
	"fmt"
	"runtime"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/shima-park/agollo"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
)

// 命名空间定义
const (
	FrameNamespace      = "frame"
	ServicesNamespace   = "services"
	ComponentsNamespace = "components"
)

// 所有支持的命名空间
var defaultNamespaces = []string{
	FrameNamespace,
	ServicesNamespace,
	ComponentsNamespace,
}

// 分析apollo配置各命名空间的key等级
const (
	analyseApolloConfigFrameKeyLevel      = 1 // frame
	analyseApolloConfigServicesKeyLevel   = 1 // services
	analyseApolloConfigComponentsKeyLevel = 2 // components
)

// 分析apollo配置各命名空间的key前缀
const (
	analyseApolloConfigFrameKeyPrefixes      = "Log,Labels" // frame
	analyseApolloConfigServicesKeyPrefixes   = "*"          // services
	analyseApolloConfigComponentsKeyPrefixes = "*"          // components
)

type ApolloConfig struct {
	Address              string // apollo-api地址, 多个地址用英文逗号连接
	AppId                string // 应用名
	AccessKey            string // 验证key, 优先级高于基础认证
	AuthBasicUser        string // 基础认证用户名
	AuthBasicPassword    string // 基础认证密码
	Cluster              string // 集群名
	AlwaysLoadFromRemote bool   // 总是从远程获取, 在远程加载失败时不会从备份文件加载
	BackupFile           string // 备份文件名
	Namespaces           string // 其他分片,多个分片名用英文逗号隔开
}

// 从viper构建apollo配置
func makeApolloConfigFromViper(vi *viper.Viper) (*ApolloConfig, error) {
	var conf ApolloConfig
	err := vi.UnmarshalKey(consts.ApolloConfigKey, &conf)
	return &conf, err
}

// 从apollo中获取配置构建viper
func makeViperFromApollo(conf *ApolloConfig) (*viper.Viper, error) {
	// 构建选项
	opts := []agollo.Option{
		agollo.AutoFetchOnCacheMiss(),                                      // 当本地缓存中namespace不存在时，尝试去apollo缓存接口去获取
		agollo.Cluster(utils.Ternary.Or(conf.Cluster, "default").(string)), // 集群名
	}
	if !conf.AlwaysLoadFromRemote {
		opts = append(opts, agollo.FailTolerantOnBackupExists()) // 从服务获取数据失败时从备份文件加载
	}
	if conf.BackupFile != "" {
		opts = append(opts, agollo.BackupFile(conf.BackupFile))
	} else if runtime.GOOS == "windows" {
		opts = append(opts, agollo.BackupFile("/nul"))
	} else {
		opts = append(opts, agollo.BackupFile("/dev/null"))
	}

	// 验证方式
	if conf.AccessKey != "" {
		opts = append(opts, agollo.AccessKey(conf.AccessKey))
	} else if conf.AuthBasicUser != "" {
		opts = append(opts,
			agollo.WithClientOptions(
				agollo.WithAccessKey("basic "+base64.StdEncoding.EncodeToString([]byte(conf.AuthBasicUser+":"+conf.AuthBasicPassword))),
				agollo.WithSignatureFunc(func(ctx *agollo.SignatureContext) agollo.Header {
					return agollo.Header{"authorization": ctx.AccessKey}
				}),
			))
	}

	namespaces := append([]string{}, defaultNamespaces...)
	if conf.Namespaces != "" {
		namespaces = append(namespaces, strings.Split(conf.Namespaces, ",")...)
	}

	// 预加载数据, 从远程或本地加载成功就不会返回错误
	opts = append(opts, agollo.PreloadNamespaces(namespaces...))

	// 构建apollo客户端
	apolloClient, err := agollo.New(conf.Address, conf.AppId, opts...)
	if err != nil {
		return nil, fmt.Errorf("初始化agollo失败: %s", err)
	}

	configs := make(map[string]interface{}, len(namespaces))
	for _, namespace := range namespaces {
		raw := apolloClient.GetNameSpace(namespace)
		data := map[string]interface{}(raw)
		switch namespace {
		case FrameNamespace:
			data = analyseApolloConfig(namespace, raw, analyseApolloConfigFrameKeyLevel, analyseApolloConfigFrameKeyPrefixes)
		case ServicesNamespace:
			data = analyseApolloConfig(namespace, raw, analyseApolloConfigServicesKeyLevel, analyseApolloConfigServicesKeyPrefixes)
		case ComponentsNamespace:
			data = analyseApolloConfig(namespace, raw, analyseApolloConfigComponentsKeyLevel, analyseApolloConfigComponentsKeyPrefixes)
		}
		configs[strings.ReplaceAll(namespace, "_", "")] = data
	}

	// 构建viper
	vi := viper.New()
	if err = vi.MergeConfigMap(configs); err != nil {
		return nil, fmt.Errorf("合并配置失败: %s", err)
	}
	return vi, nil
}

// 分析apollo配置, 它会匹配key前缀且key的层级数值为level, 然后将value转为 map[string]interface{}
func analyseApolloConfig(namespace string, raw map[string]interface{}, level int, prefixes string) map[string]interface{} {
	prefixMap := make(map[string]bool)
	for _, prefix := range strings.Split(strings.ToLower(prefixes), ",") {
		prefixMap[prefix] = true
	}

	data := make(map[string]interface{})
	for key, value := range raw {
		keys := strings.Split(key, ".")
		if len(keys) != level { // 不匹配等级
			data[key] = value
			continue
		}

		if prefixes != "*" && !prefixMap[strings.ToLower(keys[0])] { // 不匹配前缀
			data[key] = value
			continue
		}

		var doc interface{}
		err := jsoniter.UnmarshalFromString(fmt.Sprint(value), &doc)
		if err != nil {
			logger.Log.Fatal("apollo的value无法转为json", zap.String("namespace", namespace), zap.String("key", key), zap.Error(err))
		}
		data[key] = doc

	}
	return data
}
