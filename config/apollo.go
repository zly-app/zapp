/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/12/18
   Description :
-------------------------------------------------
*/

package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"

	"github.com/zly-app/zapp/config/apollo_sdk"
	"github.com/zly-app/zapp/consts"
)

const defApplicationDataType = "yaml"

type ApolloConfig struct {
	Address                 string // apollo-api地址, 多个地址用英文逗号连接
	AppId                   string // 应用名
	AccessKey               string // 验证key, 优先级高于基础认证
	AuthBasicUser           string // 基础认证用户名, 可用于nginx的基础认证扩展
	AuthBasicPassword       string // 基础认证密码
	Cluster                 string // 集群名, 默认default
	AlwaysLoadFromRemote    bool   // 总是从远程获取, 在远程加载失败时不会从备份文件加载
	BackupFile              string // 备份文件名
	ApplicationDataType     string // application命名空间下key的值的数据类型, 支持yaml,yml,toml,json
	Namespaces              string // 其他自定义命名空间, 多个命名空间用英文逗号隔开
	IgnoreNamespaceNotFound bool   // 是否忽略命名空间不存在, 无论如何设置application命名空间必须存在
	client                  *apollo_sdk.ApolloClient
}

// 从viper构建apollo配置
func makeApolloConfigFromViper(vi *viper.Viper) (*ApolloConfig, error) {
	if !vi.IsSet(consts.ApolloConfigKey) {
		return nil, errors.New("apollo配置不存在")
	}

	var conf ApolloConfig
	v := vi.Get(consts.ApolloConfigKey)
	if value, ok := v.(*ApolloConfig); ok {
		conf = *value
		if conf.client != nil {
			return &conf, nil
		}
	} else {
		err := vi.UnmarshalKey(consts.ApolloConfigKey, &conf)
		if err != nil {
			return nil, err
		}
	}

	if conf.Cluster == "" {
		conf.Cluster = os.Getenv(consts.ApolloConfigClusterFromEnvKey)
	}

	switch v := strings.ToLower(conf.ApplicationDataType); v {
	case "":
		conf.ApplicationDataType = defApplicationDataType
	case "yaml", "yml", "json", "toml":
		conf.ApplicationDataType = v
	default:
		return nil, fmt.Errorf("不支持的ApplicationDataType: %v", conf.ApplicationDataType)
	}

	ac := &apollo_sdk.ApolloClient{
		Address:                 conf.Address,
		AppId:                   conf.AppId,
		AccessKey:               conf.AccessKey,
		AuthBasicUser:           conf.AuthBasicUser,
		AuthBasicPassword:       conf.AuthBasicPassword,
		Cluster:                 conf.Cluster,
		AlwaysLoadFromRemote:    conf.AlwaysLoadFromRemote,
		BackupFile:              conf.BackupFile,
		Namespaces:              nil,
		IgnoreNamespaceNotFound: conf.IgnoreNamespaceNotFound,
	}
	if conf.Namespaces != "" {
		ac.Namespaces = strings.Split(conf.Namespaces, ",")
	}
	err := ac.Init()
	if err != nil {
		return nil, err
	}
	conf.client = ac

	vi.Set(consts.ApolloConfigKey, &conf)
	return &conf, nil
}

// 从apollo中获取配置构建viper
func makeViperFromApollo(conf *ApolloConfig) (*viper.Viper, error) {
	dataList, err := conf.client.GetNamespacesData()
	if err != nil {
		return nil, fmt.Errorf("获取apollo配置数据失败: %s", err)
	}

	configs := make(map[string]interface{}, len(dataList))
	for namespace, data := range dataList {
		err := analyseApolloConfig(configs, namespace, data.Configurations, conf.ApplicationDataType)
		if err != nil {
			return nil, fmt.Errorf("分析apollo配置数据失败: %v", err)
		}
	}

	// 构建viper
	vi := viper.New()
	if err = vi.MergeConfigMap(configs); err != nil {
		return nil, fmt.Errorf("合并apollo配置数据失败: %s", err)
	}
	return vi, nil
}

// 分析apollo配置
func analyseApolloConfig(dst map[string]interface{}, namespace string, configurations map[string]string, dType string) error {
	if namespace != apollo_sdk.ApplicationNamespace {
		dst[namespace] = configurations
		fmt.Println(fmt.Sprintf("%s: %+v", namespace, configurations))
		return nil
	}

	for k, v := range configurations {
		vi := viper.New()
		vi.SetConfigType(dType)
		err := vi.ReadConfig(strings.NewReader(v))
		if err != nil {
			return fmt.Errorf("解析数据失败 key: %v, err: %v", k, err)
		}
		data := vi.AllSettings()
		dst[k] = data
	}
	return nil
}

// 获取apollo客户端
func GetApolloClient() (*apollo_sdk.ApolloClient, error) {
	if Conf == nil {
		return nil, fmt.Errorf("config未初始化")
	}
	vi := Conf.GetViper()
	conf, err := makeApolloConfigFromViper(vi)
	if err != nil {
		return nil, err
	}
	return conf.client, nil
}
