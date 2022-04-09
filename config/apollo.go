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
	"os"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/config/apollo_sdk"
	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
)

type ApolloConfig = apollo_sdk.ApolloConfig

// 将命名空间配置的匹配key的value解析为json, 通配符语法, ? 表示一个字符, * 表示任意字符串或空字符串
var parseConfigToJsonOfMatchKeys = map[string]struct {
	matchKeys   string // 匹配key
	noMatchKeys string // 不匹配key
}{
	apollo_sdk.FrameNamespace:      {"log,flags,labels", ""},
	apollo_sdk.PluginsNamespace:    {"*", "*.*"},   // 忽略存在.的key
	apollo_sdk.ServicesNamespace:   {"*", "*.*"},   // 忽略存在.的key
	apollo_sdk.ComponentsNamespace: {"*", "*.*.*"}, // 忽略存在两个.以上的key
}

// 注册将命名空间配置的匹配key的value解析为json, 通配符语法, ? 表示一个字符, * 表示任意字符串或空字符串
//
// matchKeys 表示匹配key的值作为json解析, 多个key用英文逗号连接, 多个key只要匹配其中一个则满足条件.
// noMatchKeys 表示匹配key的值不作为json解析, 多个key用英文逗号连接, 多个key只要匹配其中一个则满足条件.
// 如果设置了 noMatchKeys 但是没有设置 matchKeys, 表示如果没有满足 noMatchKeys 则将值作为json解析.
func RegistryParseConfigToJsonOfMatchKeys(namespace string, matchKeys, noMatchKeys string) {
	parseConfigToJsonOfMatchKeys[namespace] = struct {
		matchKeys   string
		noMatchKeys string
	}{matchKeys, noMatchKeys}
}

// 从viper构建apollo配置
func makeApolloConfigFromViper(vi *viper.Viper) (*ApolloConfig, error) {
	var conf ApolloConfig
	err := vi.UnmarshalKey(consts.ApolloConfigKey, &conf)
	return &conf, err
}

// 从apollo中获取配置构建viper
func makeViperFromApollo(conf *ApolloConfig) (*viper.Viper, error) {
	if conf.Cluster == "" {
		conf.Cluster = os.Getenv(consts.ApolloConfigClusterFromEnvKey)
	}

	dataList, err := conf.GetNamespacesData()
	if err != nil {
		return nil, fmt.Errorf("获取apollo配置数据失败: %s", err)
	}

	configs := make(map[string]interface{}, len(dataList))
	for namespace, data := range dataList {
		d := analyseApolloConfig(namespace, data.Configurations)
		configs[namespace] = d
	}

	// 构建viper
	vi := viper.New()
	if err = vi.MergeConfigMap(configs); err != nil {
		return nil, fmt.Errorf("合并配置失败: %s", err)
	}
	return vi, nil
}

// 分析apollo配置, 然后匹配key的value转为 map[string]interface{}
func analyseApolloConfig(namespace string, configurations map[string]string) map[string]interface{} {
	matchKey, ok := parseConfigToJsonOfMatchKeys[namespace]
	if !ok || (matchKey.matchKeys == "" && matchKey.noMatchKeys == "") { // 没有设置匹配key
		result := make(map[string]interface{})
		for k, v := range configurations {
			result[k] = v
		}
		return result
	}

	rawMatchKeys, rawNoMatchKeys := matchKey.matchKeys, matchKey.noMatchKeys
	var matchKeys, noMatchKeys []string
	if rawMatchKeys != "" {
		matchKeys = strings.Split(rawMatchKeys, ",")
	}
	if rawNoMatchKeys != "" {
		noMatchKeys = strings.Split(rawNoMatchKeys, ",")
	}

	data := make(map[string]interface{})
	for key, value := range configurations {
		if (len(matchKeys) > 0 && !utils.Text.IsMatchWildcardAny(strings.ToLower(key), matchKeys...)) ||
			(len(noMatchKeys) > 0 && utils.Text.IsMatchWildcardAny(strings.ToLower(key), noMatchKeys...)) {
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

// 获取apollo客户端
func GetApolloClient() (*ApolloConfig, error) {
	if Conf == nil {
		return nil, fmt.Errorf("config未初始化")
	}
	vi := Conf.GetViper()
	return makeApolloConfigFromViper(vi)
}
