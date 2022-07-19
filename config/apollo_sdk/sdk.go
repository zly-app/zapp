/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/22
   Description :
-------------------------------------------------
*/

package apollo_sdk

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/zly-app/zapp/logger"
)

// apollo获取配置api
// https://github.com/ctripcorp/apollo/wiki/%E5%85%B6%E5%AE%83%E8%AF%AD%E8%A8%80%E5%AE%A2%E6%88%B7%E7%AB%AF%E6%8E%A5%E5%85%A5%E6%8C%87%E5%8D%97
// https://www.apolloconfig.com/#/zh/usage/other-language-client-user-guide?id=_13-%e9%80%9a%e8%bf%87%e4%b8%8d%e5%b8%a6%e7%bc%93%e5%ad%98%e7%9a%84http%e6%8e%a5%e5%8f%a3%e4%bb%8eapollo%e8%af%bb%e5%8f%96%e9%85%8d%e7%bd%ae
// {config_server_url}/configs/{appId}/{clusterName}/{namespaceName}?releaseKey={releaseKey}&ip={clientIp}
const ApolloGetNamespaceDataApiUrl = "/configs/%s/%s/%s?releaseKey=%s&ip=%s"

const (
	// apollo获取通知api
	// {config_server_url}/notifications/v2?appId={appId}&cluster={clusterName}&notifications={notifications}
	ApolloWatchNamespaceChangedApiUrl = "/notifications/v2?appId=%s&cluster=%s&notifications=%s"
)

var (
	// http请求超时
	HttpReqTimeout = time.Second * 3
	// http请求通知超时
	HttpReqNotificationTimeout = time.Second * 65
)

// 错误状态码描述
var errStatusCodesDescribe = map[int]string{
	400: "客户端传入参数的错误",
	401: "客户端未授权或认证失败",
	404: "命名空间数据不存在",
	405: "接口访问的Method不正确",
	500: "服务内部错误",
}

// 默认命名空间, 不会加上 NamespacePrefix
const ApplicationNamespace = "application"

type ApolloClient struct {
	Address                 string   // apollo-api地址, 多个地址用英文逗号连接
	AppId                   string   // 应用名
	AccessKey               string   // 验证key, 优先级高于基础认证
	AuthBasicUser           string   // 基础认证用户名, 可用于nginx的基础认证扩展
	AuthBasicPassword       string   // 基础认证密码
	Cluster                 string   // 集群名, 默认default
	AlwaysLoadFromRemote    bool     // 总是从远程获取, 在远程加载失败时不会从备份文件加载
	BackupFile              string   // 备份文件名
	Namespaces              []string // 其他自定义命名空间
	IgnoreNamespaceNotFound bool     // 是否忽略命名空间不存在
	cache                   MultiNamespaceData
}

type (
	// 命名空间数据
	NamespaceData = struct {
		AppId          string            `json:"appId"`
		Cluster        string            `json:"cluster"`
		Namespace      string            `json:"namespaceName"`
		Configurations map[string]string `json:"configurations"`
		ReleaseKey     string            `json:"releaseKey"`
	}
	// 多个命名空间数据
	MultiNamespaceData = map[string]*NamespaceData
	// 通知参数
	NotificationParam struct {
		NamespaceName  string `json:"namespaceName"`
		NotificationId int    `json:"notificationId"`
	}
	// 通知结果数据
	NotificationRsp struct {
		NamespaceName  string `json:"namespaceName"`
		NotificationId int    `json:"notificationId"`
	}
)

func (a *ApolloClient) clientIP() string {
	return ""
}

func (a *ApolloClient) Init() error {
	a.cache = make(MultiNamespaceData)
	return nil
}

// 获取所有命名空间的数据
func (a *ApolloClient) GetNamespacesData() (MultiNamespaceData, error) {
	namespaces := append([]string{ApplicationNamespace}, a.Namespaces...)
	// 允许从本地备份获取
	if a.isAllowLoadFromBackupFile() {
		backupData, err := a.loadDataFromBackupFile()
		if err != nil {
			logger.Log.Error("从本地加载配置失败", zap.Error(err))
		} else {
			a.writeCache(backupData)
		}
	}

	// 退出之前保存当前已存在数据
	defer a.saveDataToBackupFile()

	// 遍历获取
	for _, namespace := range namespaces {
		// 从远程获取数据
		remoteData, _, err := a.loadNamespaceDataFromRemote(namespace)
		if err == nil { // 成功拿到则覆盖数据
			a.cache[namespace] = remoteData
			continue
		}

		// 如果总是从远程获取则返回错误
		if a.AlwaysLoadFromRemote {
			return nil, fmt.Errorf("从远程获取命名空间<%s>的数据失败: %s", namespace, err)
		}

		logger.Log.Error("从远程获取配置失败", zap.String("namespace", namespace), zap.Error(err))
		_, ok := a.cache[namespace]
		if !ok {
			return nil, fmt.Errorf("本地命名空间<%s>的数据不存在", namespace)
		}
	}

	return a.cache, nil
}

// 从远程加载命名空间数据
func (a *ApolloClient) loadNamespaceDataFromRemote(namespace string) (data *NamespaceData, changed bool, err error) {
	// 检查配置
	if a.Address == "" {
		return nil, false, errors.New("apollo的address是空的")
	}
	if a.AppId == "" {
		return nil, false, errors.New("apollo的appid是空的")
	}
	if a.Cluster == "" {
		a.Cluster = "default"
	}

	cacheData, hasCache := a.cache[namespace]

	var requestUri string
	if hasCache {
		requestUri = fmt.Sprintf(ApolloGetNamespaceDataApiUrl, a.AppId, a.Cluster, namespace, cacheData.ReleaseKey, a.clientIP())
	} else {
		requestUri = fmt.Sprintf(ApolloGetNamespaceDataApiUrl, a.AppId, a.Cluster, namespace, "", a.clientIP())
	}

	// 构建请求体
	// 超时
	ctx, cancel := context.WithTimeout(context.Background(), HttpReqTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", a.Address+requestUri, nil)
	if err != nil {
		return nil, false, err
	}
	a.officialSignature(req) // 认证

	// 请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound && a.IgnoreNamespaceNotFound && namespace != ApplicationNamespace { // 命名空间不存在
			empty := &NamespaceData{
				AppId:          a.AppId,
				Cluster:        a.Cluster,
				Namespace:      namespace,
				Configurations: make(map[string]string),
			}
			return empty, true, nil // 视为空配置数据
		}
		if resp.StatusCode == http.StatusNotModified { // 未改变
			return cacheData, false, nil
		}

		desc, ok := errStatusCodesDescribe[resp.StatusCode]
		if !ok {
			desc = "未知错误"
		}
		return nil, false, fmt.Errorf("收到错误码: %d: %s", resp.StatusCode, desc)
	}

	// 解码
	var result NamespaceData
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, false, fmt.Errorf("解码失败: %v", err)
	}
	if result.Configurations == nil {
		result.Configurations = make(map[string]string)
	}
	return &result, true, nil
}

/*获取命名空间数据
  如果 oldData 为 nil 会直接获取数据
  如果 oldData.ReleaseKey 不为空则检查是否会改变了
*/
func (a *ApolloClient) GetNamespaceData(namespace string) (oldData, newData *NamespaceData, changed bool, err error) {
	oldData = a.cache[namespace]
	if oldData == nil {
		oldData = &NamespaceData{
			AppId:          a.AppId,
			Cluster:        a.Cluster,
			Namespace:      namespace,
			Configurations: make(map[string]string),
		}
	}

	newData, changed, err = a.loadNamespaceDataFromRemote(namespace)
	if changed {
		a.cache[namespace] = newData
		a.saveDataToBackupFile()
	}
	return
}

/*等待通知
  如果数据未被改变, 此方法会导致挂起直到超时或被改变
*/
func (a *ApolloClient) WaitNotification(ctx context.Context, param []*NotificationParam) ([]*NotificationRsp, error) {
	if len(param) == 0 {
		return nil, nil
	}
	// 检查配置
	if a.Address == "" {
		return nil, errors.New("apollo的address是空的")
	}
	if a.AppId == "" {
		return nil, errors.New("apollo的appid是空的")
	}
	if a.Cluster == "" {
		a.Cluster = "default"
	}

	paramData, _ := json.Marshal(param)
	requestUri := fmt.Sprintf(ApolloWatchNamespaceChangedApiUrl, a.AppId, a.Cluster, url.QueryEscape(string(paramData)))

	// 超时
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, HttpReqNotificationTimeout)
	defer cancel()

	// 构建请求体
	req, err := http.NewRequestWithContext(ctx, "GET", a.Address+requestUri, nil)
	if err != nil {
		return nil, err
	}
	a.officialSignature(req) // 认证

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if err == context.Canceled { // 被主动取消
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotModified { // 状态未改变
			return nil, nil
		}

		desc, ok := errStatusCodesDescribe[resp.StatusCode]
		if !ok {
			desc = "未知错误"
		}
		return nil, fmt.Errorf("收到错误码: %d: %s", resp.StatusCode, desc)
	}

	// 解码
	var result []*NotificationRsp
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("解码失败: %v", err)
	}
	return result, nil
}

// 保存数据到备份文件
func (a *ApolloClient) saveDataToBackupFile() {
	if len(a.cache) == 0 || a.BackupFile == "" {
		return
	}

	bs, err := yaml.Marshal(a.cache)
	if err == nil {
		err = ioutil.WriteFile(a.BackupFile, bs, 0644)
	}
	if err != nil {
		logger.Log.Error("备份配置文件失败", zap.Error(err))
	}
}

// 从备份文件加载数据
func (a *ApolloClient) loadDataFromBackupFile() (MultiNamespaceData, error) {
	if a.BackupFile == "" {
		return nil, nil
	}

	bs, err := ioutil.ReadFile(a.BackupFile)
	if err != nil {
		return nil, err
	}

	var result MultiNamespaceData
	err = yaml.Unmarshal(bs, &result)
	return result, err
}

// 写入缓存
func (a *ApolloClient) writeCache(data MultiNamespaceData) {
	for k, v := range data {
		a.cache[k] = v
	}
}

// 是否允许从本地备份获取
func (a *ApolloClient) isAllowLoadFromBackupFile() bool {
	return !a.AlwaysLoadFromRemote && a.BackupFile != "" // 不总是从远程获取 并且 存在备份文件
}

// 官方签名
func (a *ApolloClient) officialSignature(req *http.Request) {
	if a.AccessKey != "" {
		timestamp := fmt.Sprintf("%v", time.Now().UnixNano()/int64(time.Millisecond))
		stringToSign := timestamp + "\n" + req.URL.RequestURI()
		key := []byte(a.AccessKey)
		mac := hmac.New(sha1.New, key)
		_, _ = mac.Write([]byte(stringToSign))
		signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		req.Header.Add("Authorization", fmt.Sprintf("Apollo %s:%s", a.AppId, signature))
		req.Header.Add("Timestamp", timestamp)
		return
	}

	if a.AuthBasicUser != "" {
		req.Header.Add("Authorization", fmt.Sprintf("basic %s", base64.StdEncoding.EncodeToString([]byte(a.AuthBasicUser+":"+a.AuthBasicPassword))))
		return
	}
}
