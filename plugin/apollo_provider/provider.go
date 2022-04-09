package apollo_provider

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/config/apollo_sdk"
	"github.com/zly-app/zapp/core"
)

// 观察失败等待时间
var WatchErrWaitTime = time.Second * 5

type ApolloProvider struct {
	app    core.IApp
	client *config.ApolloConfig

	watchNamespaces       map[string]int          // 观察的命名空间, value为notificationID
	namespaceCallbackList map[string]KeyCallbacks // 命名空间回调列表

	watchCtx       context.Context
	watchCtxCancel context.CancelFunc

	// 用于锁 watchNamespaces, watchNamespaceList
	mx sync.Mutex
}

// key回调列表
type KeyCallbacks map[string][]core.ConfigWatchProviderCallback

func (p *ApolloProvider) Inject(a ...interface{}) {}
func (p *ApolloProvider) Start() error {
	go p.startWatchNamespace()
	return nil
}
func (p *ApolloProvider) Close() error {
	p.watchCtxCancel()
	return nil
}

func NewApolloProvider(app core.IApp) *ApolloProvider {
	client, err := config.GetApolloClient()
	if err != nil {
		app.Fatal("获取客户端失败", zap.Error(err))
	}
	p := &ApolloProvider{
		app:                   app,
		client:                client,
		watchNamespaces:       make(map[string]int),
		namespaceCallbackList: make(map[string]KeyCallbacks),
	}
	p.watchCtx, p.watchCtxCancel = context.WithCancel(app.BaseContext())
	return p
}

// 获取
func (p *ApolloProvider) Get(groupName, keyName string) ([]byte, error) {
	_, data, _, err := p.client.GetNamespaceData(groupName)
	if err != nil {
		return nil, err
	}
	value, ok := data.Configurations[keyName]
	if !ok {
		return nil, fmt.Errorf("配置数据不存在 groupName: %s, keyName: %s", groupName, keyName)
	}
	return []byte(value), nil
}

// watch
func (p *ApolloProvider) Watch(groupName, keyName string, callback core.ConfigWatchProviderCallback) error {
	_, data, _, err := p.client.GetNamespaceData(groupName)
	if err != nil {
		return err
	}
	_, ok := data.Configurations[keyName]
	if !ok {
		return fmt.Errorf("配置数据不存在 groupName: %s, keyName: %s", groupName, keyName)
	}

	// 添加观察命名空间
	p.addWatchNamespace(groupName)
	// 添加回调
	p.addCallback(groupName, keyName, callback)

	return nil
}

// 添加观察命名空间
func (p *ApolloProvider) addWatchNamespace(namespace string) {
	p.mx.Lock()
	defer p.mx.Unlock()

	_, ok := p.watchNamespaces[namespace]
	if !ok {
		p.watchNamespaces[namespace] = 0
	}
}

// 添加回调
func (p *ApolloProvider) addCallback(groupName, keyName string, callback core.ConfigWatchProviderCallback) {
	p.mx.Lock()
	defer p.mx.Unlock()

	keyCallbacks, ok := p.namespaceCallbackList[groupName]
	if !ok {
		keyCallbacks = make(KeyCallbacks, 1)
		p.namespaceCallbackList[groupName] = keyCallbacks
	}
	keyCallbacks[keyName] = append(keyCallbacks[keyName], callback)
}

// 开始观察命名空间
func (p *ApolloProvider) startWatchNamespace() {
	for {
		select {
		case <-p.watchCtx.Done():
			return
		default:
			param := p.makeNotificationParam()
			rsp, err := p.client.WaitNotification(p.watchCtx, param)
			if err != nil {
				time.Sleep(WatchErrWaitTime)
				continue
			}

			// 解析通知结果
			p.parseNotificationRsp(rsp)
		}
	}
}

// 构建通知param
func (p *ApolloProvider) makeNotificationParam() []*apollo_sdk.NotificationParam {
	p.mx.Lock()
	param := make([]*apollo_sdk.NotificationParam, 0, len(p.watchNamespaces))
	for k, nid := range p.watchNamespaces {
		param = append(param, &apollo_sdk.NotificationParam{
			NamespaceName:  k,
			NotificationId: nid,
		})
	}
	p.mx.Unlock()
	return param
}

// 解析通知结果
func (p *ApolloProvider) parseNotificationRsp(rsp []*apollo_sdk.NotificationRsp) {
	if len(rsp) == 0 {
		return
	}

	p.mx.Lock()
	defer p.mx.Unlock()
	for _, v := range rsp {
		p.watchNamespaces[v.NamespaceName] = v.NotificationId
		go p.ReReqNamespaceData(v.NamespaceName)
	}
}

// 重新请求命名空间数据
func (p *ApolloProvider) ReReqNamespaceData(namespace string) {
	oldData, newData, changed, err := p.client.GetNamespaceData(namespace)
	if err != nil {
		p.app.Error("重新请求apollo命名空间数据失败", zap.String("namespace", namespace), zap.Error(err))
		return
	}
	if !changed {
		return
	}

	p.mx.Lock()
	defer p.mx.Unlock()

	// 获取回调函数列表, 遍历回调
	keyCallbacks, ok := p.namespaceCallbackList[namespace]
	if !ok {
		return
	}
	for key, callbacks := range keyCallbacks {
		oldVale := oldData.Configurations[key]
		newValue := newData.Configurations[key]
		for _, fn := range callbacks {
			go fn(namespace, key, []byte(oldVale), []byte(newValue))
		}
	}
}
