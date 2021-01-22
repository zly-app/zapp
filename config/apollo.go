/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/config/apollo_sdk"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/logger"
)

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

type ApolloConfig = apollo_sdk.ApolloConfig

// 从viper构建apollo配置
func makeApolloConfigFromViper(vi *viper.Viper) (*ApolloConfig, error) {
	var conf ApolloConfig
	err := vi.UnmarshalKey(consts.ApolloConfigKey, &conf)
	return &conf, err
}

// 从apollo中获取配置构建viper
func makeViperFromApollo(conf *ApolloConfig) (*viper.Viper, error) {
	data, err := conf.GetNamespacesData()
	if err != nil {
		return nil, fmt.Errorf("获取apollo配置数据失败: %s", err)
	}

	configs := make(map[string]interface{}, len(data))
	for namespace, raw := range data {
		data := map[string]interface{}(raw)
		switch namespace {
		case apollo_sdk.FrameNamespace:
			data = analyseApolloConfig(namespace, raw, analyseApolloConfigFrameKeyLevel, analyseApolloConfigFrameKeyPrefixes)
		case apollo_sdk.ServicesNamespace:
			data = analyseApolloConfig(namespace, raw, analyseApolloConfigServicesKeyLevel, analyseApolloConfigServicesKeyPrefixes)
		case apollo_sdk.ComponentsNamespace:
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
		err := json.Unmarshal([]byte(fmt.Sprint(value)), &doc)
		if err != nil {
			logger.Log.Fatal("apollo的value无法转为json", zap.String("namespace", namespace), zap.String("key", key), zap.Error(err))
		}
		data[key] = doc

	}
	return data
}
