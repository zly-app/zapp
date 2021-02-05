/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/22
   Description :
-------------------------------------------------
*/

package apollo_sdk

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/logger"
)

// apollo获取配置api
// https://github.com/ctripcorp/apollo/wiki/%E5%85%B6%E5%AE%83%E8%AF%AD%E8%A8%80%E5%AE%A2%E6%88%B7%E7%AB%AF%E6%8E%A5%E5%85%A5%E6%8C%87%E5%8D%97
const ApolloGetNamespaceDataApiUrl = "/configfiles/json/%s/%s/%s" //  {config_server_url}/configfiles/json/{appId}/{clusterName}/{namespaceName}

// 命名空间定义
const (
	FrameNamespace      = "frame"
	ServicesNamespace   = "services"
	ComponentsNamespace = "components"
)

// 默认需要的的命名空间
var defaultNamespaces = []string{
	FrameNamespace,
	ServicesNamespace,
	ComponentsNamespace,
}

// 错误状态码描述
var errStatusCodesDescribe = map[int]string{
	400: "客户端传入参数的错误",
	401: "客户端未授权或认证失败",
	404: "网络错误或命名空间数据不存在",
	405: "接口访问的Method不正确",
	500: "服务内部错误",
}

type ApolloConfig struct {
	Address              string // apollo-api地址, 多个地址用英文逗号连接
	AppId                string // 应用名
	AccessKey            string // 验证key, 优先级高于基础认证
	AuthBasicUser        string // 基础认证用户名, 可用于nginx的基础认证扩展
	AuthBasicPassword    string // 基础认证密码
	Cluster              string // 集群名, 默认default
	AlwaysLoadFromRemote bool   // 总是从远程获取, 在远程加载失败时不会从备份文件加载
	BackupFile           string // 备份文件名
	Namespaces           string // 其他自定义命名空间, 多个命名空间用英文逗号隔开
}

type (
	// 命名空间数据
	NamespaceData = map[string]interface{}
	// 多个命名空间数据
	MultiNamespaceData = map[string]NamespaceData
)

// 获取指定命名空间的数据
func (a *ApolloConfig) GetNamespacesData() (MultiNamespaceData, error) {
	namespaces := append([]string{}, defaultNamespaces...)
	if a.Namespaces != "" {
		namespaces = append(namespaces, strings.Split(a.Namespaces, ",")...)
	}

	data := make(MultiNamespaceData, len(namespaces))
	result := make(MultiNamespaceData, len(namespaces))

	// 允许从本地获取
	if a.isAllowBackupFile() {
		backupData, err := a.loadDataFromBackupFile()
		if err == nil {
			a.overrideMultiNamespaceData(data, backupData)
		}
	}

	// 退出之前保存当前已存在数据
	defer a.saveDataToBackupFile(data)

	// 遍历获取
	for _, namespace := range namespaces {
		// 从远程获取数据
		raw, err := a.getNamespaceDataFromRemote(namespace)

		// 成功拿到则覆盖数据
		if err == nil {
			if raw == nil {
				raw = make(NamespaceData, 0)
			}
			data[namespace] = raw
			result[namespace] = raw
			continue
		}

		// 如果不使用备份文件或无历史数据则直接返回错误
		if !a.isAllowBackupFile() || data[namespace] == nil {
			return nil, fmt.Errorf("获取命名空间<%s>的数据失败: %s", namespace, err)
		}

		result[namespace] = data[namespace]
	}

	return result, nil
}

// 从远程获取命名空间数据
func (a *ApolloConfig) getNamespaceDataFromRemote(namespace string) (NamespaceData, error) {
	// 检查配置
	if a.Address == "" {
		return nil, errors.New("apollo的address是空的")
	}
	if a.AppId == "" {
		return nil, errors.New("apollo的appid是空的")
	}
	cluster := a.Cluster
	if cluster == "" {
		cluster = "default"
	}

	// 构建请求体
	requestUri := fmt.Sprintf(ApolloGetNamespaceDataApiUrl, a.AppId, cluster, namespace)
	req, err := http.NewRequest("GET", a.Address+requestUri, nil)
	if err != nil {
		return nil, err
	}

	// 认证
	if a.AccessKey != "" {
		timestamp := fmt.Sprintf("%v", time.Now().UnixNano()/int64(time.Millisecond))
		signature := a.officialSignature(timestamp, requestUri, a.AccessKey)
		req.Header.Add("Authorization", fmt.Sprintf("Apollo %s:%s", a.AppId, signature))
		req.Header.Add("Timestamp", timestamp)
	} else if a.AuthBasicUser != "" {
		req.Header.Add("Authorization", fmt.Sprintf("basic %s", base64.StdEncoding.EncodeToString([]byte(a.AuthBasicUser+":"+a.AuthBasicPassword))))
	}

	// 请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 404时尝试检查是否namespace不存在
	if resp.StatusCode == http.StatusNotFound {
		var result struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			if result.Error == "Not Found" {
				logger.Log.Warn("命名空间不存在", zap.String("namespace", namespace))
				return NamespaceData{}, nil
			}
		}
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		desc, ok := errStatusCodesDescribe[resp.StatusCode]
		if !ok {
			desc = "未知错误"
		}
		return nil, fmt.Errorf("收到错误码: %d: %s", resp.StatusCode, desc)
	}

	// 解码
	var result NamespaceData
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

// 保存数据到备份文件
func (a *ApolloConfig) saveDataToBackupFile(data MultiNamespaceData) error {
	if len(data) == 0 || a.BackupFile == "" {
		return nil
	}

	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(a.BackupFile, bs, 0644)
}

// 从备份文件加载数据
func (a *ApolloConfig) loadDataFromBackupFile() (MultiNamespaceData, error) {
	if a.BackupFile == "" {
		return nil, nil
	}

	bs, err := ioutil.ReadFile(a.BackupFile)
	if err != nil {
		return nil, err
	}

	var result MultiNamespaceData
	err = json.Unmarshal(bs, &result)
	return result, err
}

// 将datab的数据覆盖dataa
func (a *ApolloConfig) overrideMultiNamespaceData(dataa, datab MultiNamespaceData) {
	for name, data := range datab {
		dataa[name] = data
	}
}

func (a *ApolloConfig) isAllowBackupFile() bool {
	return !a.AlwaysLoadFromRemote && a.BackupFile != ""
}

func (a *ApolloConfig) officialSignature(timestamp, url, accessKey string) string {
	stringToSign := timestamp + "\n" + url
	key := []byte(accessKey)
	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
