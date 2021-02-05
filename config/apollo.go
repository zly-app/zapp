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

type ApolloConfig = apollo_sdk.ApolloConfig

// 将命名空间配置的指定等级及主key的value解析为json, *匹配任何主key
var parseConfigToJsonOfPrimaryKeys = map[string]struct {
	primaryKey string
	keyLevel   int
}{
	apollo_sdk.FrameNamespace:      {"log,labels", 1},
	apollo_sdk.ServicesNamespace:   {"*", 1},
	apollo_sdk.ComponentsNamespace: {"*", 2},
}

// 注册将命名空间配置的哪些key解析为json, *匹配任何主key
func RegistryParseConfigToJsonOfPrimaryKey(namespace, primaryKey string, keyLevel int) {
	parseConfigToJsonOfPrimaryKeys[namespace] = struct {
		primaryKey string
		keyLevel   int
	}{primaryKey: primaryKey, keyLevel: keyLevel}
}

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
		d := map[string]interface{}(raw)
		d = analyseApolloConfig(namespace, d)
		configs[namespace] = d
	}

	// 构建viper
	vi := viper.New()
	if err = vi.MergeConfigMap(configs); err != nil {
		return nil, fmt.Errorf("合并配置失败: %s", err)
	}
	return vi, nil
}

// 分析apollo配置, 它会匹配key前缀且key的层级数值为level, 然后将value转为 map[string]interface{}
func analyseApolloConfig(namespace string, raw map[string]interface{}) map[string]interface{} {
	primaryKeys, ok := parseConfigToJsonOfPrimaryKeys[namespace]
	if !ok {
		return raw
	}

	prefixMap := make(map[string]bool)
	for _, prefix := range strings.Split(strings.ToLower(primaryKeys.primaryKey), ",") {
		prefixMap[prefix] = true
	}

	data := make(map[string]interface{})
	for key, value := range raw {
		keys := strings.Split(key, ".")
		if len(keys) != primaryKeys.keyLevel { // 不匹配等级
			data[key] = value
			continue
		}

		if primaryKeys.primaryKey != "*" && !prefixMap[strings.ToLower(keys[0])] { // 不匹配主key
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
