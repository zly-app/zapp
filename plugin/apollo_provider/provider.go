package apollo_provider

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
)

type ApolloProvider struct {
	app    core.IApp
	client *config.ApolloConfig
}

func (p *ApolloProvider) Inject(a ...interface{}) {}
func (p *ApolloProvider) Start() error            { return nil }
func (p *ApolloProvider) Close() error            { return nil }

func NewApolloProvider(app core.IApp) *ApolloProvider {
	client, err := config.GetApolloClient()
	if err != nil {
		app.Fatal("获取客户端失败", zap.Error(err))
	}
	p := &ApolloProvider{
		app:    app,
		client: client,
	}
	return p
}

// 获取
func (p *ApolloProvider) Get(groupName, keyName string) ([]byte, error) {
	data, _, err := p.client.GetNamespaceData(groupName)
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
func (p *ApolloProvider) Watch(groupName, keyName string,
	callback core.ConfigWatchProviderCallback) error {
	return nil
}
