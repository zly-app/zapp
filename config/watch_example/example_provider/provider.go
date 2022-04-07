package example_provider

import (
	"fmt"
	"sync"
	"time"

	"github.com/zly-app/zapp/core"
)

// 示例提供者
type ExampleProvider struct {
	app  core.IApp
	data map[string]map[string][]byte // 表示组下的key的数据
	mx   sync.Mutex
}

// 实现必须的函数
func (t *ExampleProvider) Inject(a ...interface{}) {}
// 实现必须的函数
func (t *ExampleProvider) Start() error            { return nil }
// 实现必须的函数
func (t *ExampleProvider) Close() error            { return nil }

func NewExamplePlugin(app core.IApp) *ExampleProvider {
	p := &ExampleProvider{
		app:  app,
		data: make(map[string]map[string][]byte),
	}

	// 下面的代码是填充一些默认数据
	data := map[string]map[string][]byte{
		"group_name": {
			"key_name": []byte("1"),
		},
	}
	p.data = data
	// 上面的代码是填充一些默认数据

	return p
}

// 获取
func (t *ExampleProvider) Get(groupName, keyName string) ([]byte, error) {
	t.mx.Lock()
	defer t.mx.Unlock()

	g, ok := t.data[groupName]
	if !ok {
		return nil, fmt.Errorf("not found group: %s", groupName)
	}

	data, ok := g[keyName]
	if !ok {
		return nil, fmt.Errorf("not found key: %s.%s", groupName, keyName)
	}

	return data, nil
}

// 设置
func (t *ExampleProvider) Set(groupName, keyName string, data []byte) error {
	t.mx.Lock()
	defer t.mx.Unlock()

	g, ok := t.data[groupName]
	if !ok {
		g = make(map[string][]byte)
		t.data[groupName] = g
	}

	g[keyName] = data
	return nil
}

// watch
func (t *ExampleProvider) Watch(groupName, keyName string,
	callback core.ConfigWatchProviderCallback) error {
	go func(groupName, keyName string, callback core.ConfigWatchProviderCallback) {
		for {
			time.Sleep(time.Second * 2)
			t.mx.Lock()
			oldData := t.getData(groupName, keyName)
			newData := make([]byte, len(oldData)+1)
			copy(newData, oldData)
			newData[len(newData)-1] = 't'
			t.data[groupName][keyName] = newData
			t.mx.Unlock()
			callback(groupName, keyName, oldData, newData)
		}
	}(groupName, keyName, callback)
	return nil
}

func (t *ExampleProvider) getData(groupName, keyName string) []byte {
	g, ok := t.data[groupName]
	if !ok {
		return nil
	}
	return g[keyName]
}
