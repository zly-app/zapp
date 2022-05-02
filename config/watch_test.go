package config

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/zly-app/zapp/core"
)

// 测试提供者
type TestWatchProvider struct {
	data map[string]map[string][]byte // 表示组下的key的数据
	mx   sync.Mutex
}

// 创建测试提供者
func NewTestWatchProvider(data map[string]map[string][]byte) *TestWatchProvider {
	p := &TestWatchProvider{
		data: data,
	}
	return p
}
func (t *TestWatchProvider) Get(groupName, keyName string) ([]byte, error) {
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
func (t *TestWatchProvider) Set(groupName, keyName string, data []byte) error {
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
func (t *TestWatchProvider) Watch(groupName, keyName string, callback core.ConfigWatchProviderCallback) error {
	go func(groupName, keyName string, callback core.ConfigWatchProviderCallback) {
		for {
			time.Sleep(time.Second * 2)
			t.mx.Lock()
			oldData := t.getData(groupName, keyName)
			newData := make([]byte, len(oldData)*2)
			copy(newData, oldData)
			copy(newData[len(oldData):], oldData)
			t.data[groupName][keyName] = newData
			t.mx.Unlock()
			callback(groupName, keyName, oldData, newData)
		}
	}(groupName, keyName, callback)
	return nil
}
func (t *TestWatchProvider) getData(groupName, keyName string) []byte {
	g, ok := t.data[groupName]
	if !ok {
		return nil
	}
	return g[keyName]
}

var testProvider *TestWatchProvider

func init() {
	p := NewTestWatchProvider(map[string]map[string][]byte{})
	testProvider = p
	RegistryConfigWatchProvider("test", p)
	SetDefaultConfigWatchProvider(p)
}

func TestSDK(t *testing.T) {
	testGroupName := "g1"
	testKeyName := "k1"

	err := testProvider.Set(testGroupName, testKeyName, []byte("1"))
	require.Nil(t, err)

	keyObj, err := newWatchKeyObject(testGroupName, testKeyName)
	require.Nil(t, err)

	// 获取原始数据
	y1 := keyObj.GetString()
	require.Equal(t, "1", y1)

	// 转为 int 值
	y2 := keyObj.GetInt()
	require.Equal(t, 1, y2)

	// 转为 boolean 值
	y3 := keyObj.GetBool()
	require.Equal(t, true, y3)

	// 检查复合预期
	b1 := keyObj.Expect("1")
	require.Equal(t, true, b1)
	b2 := keyObj.Expect(1)
	require.Equal(t, true, b2)
	b3 := keyObj.Expect(true)
	require.Equal(t, true, b3)
}

func TestWatch(t *testing.T) {
	testGroupName := "g2"
	testKeyName := "k2"

	err := testProvider.Set(testGroupName, testKeyName, []byte("2"))
	require.Nil(t, err)

	keyObj, err := newWatchKeyObject(testGroupName, testKeyName)
	require.Nil(t, err)

	var isCallback bool
	keyObj.AddCallback(func(k core.IConfigWatchKeyObject, oldData, newData []byte) {
		isCallback = true
		require.Equal(t, "", string(oldData))
		require.Equal(t, "2", string(newData))
		require.Equal(t, keyObj, k)
	})
	require.Nil(t, err)

	time.Sleep(time.Millisecond * 100)
	require.True(t, isCallback)
	require.Equal(t, "2", keyObj.GetString())
}

func TestExpect(t *testing.T) {
	testGroupName := "g3"
	testKeyName := "k3"

	err := testProvider.Set(testGroupName, testKeyName, []byte("1"))
	require.Nil(t, err)

	keyObj, err := newWatchKeyObject(testGroupName, testKeyName)
	require.Nil(t, err)

	var tests = []struct {
		expect interface{}
		result bool
	}{
		{[]byte("1"), true},
		{"1", true},
		{[]rune("1"), true},
		{true, true},
		{1, true},
		{int8(1), true},
		{int16(1), true},
		{int32(1), true},
		{int64(1), true},
		{uint(1), true},
		{uint8(1), true},
		{uint16(1), true},
		{uint32(1), true},
		{uint64(1), true},
		{float32(1), true},
		{float64(1), true},

		{[]byte("2"), false},
		{"2", false},
		{[]rune("2"), false},
		{false, false},
		{2, false},
		{int8(2), false},
		{int16(2), false},
		{int32(2), false},
		{int64(2), false},
		{uint(2), false},
		{uint8(2), false},
		{uint16(2), false},
		{uint32(2), false},
		{uint64(2), false},
		{float32(2), false},
		{float64(2), false},
	}

	for _, test := range tests {
		result := keyObj.Expect(test.expect)
		require.Equal(t, result, test.result)
	}
}

func TestConvert(t *testing.T) {
	testGroupName := "g4"
	testKeyName := "k4"

	err := testProvider.Set(testGroupName, testKeyName, []byte("4"))
	require.Nil(t, err)

	keyObj, err := newWatchKeyObject(testGroupName, testKeyName)
	require.Nil(t, err)

	require.Equal(t, []byte("4"), keyObj.GetData())
	require.Equal(t, "4", keyObj.GetString())
	require.Equal(t, false, keyObj.GetBool())
	require.Equal(t, true, keyObj.GetBool(true))
	require.Equal(t, 4, keyObj.GetInt())
	require.Equal(t, int8(4), keyObj.GetInt8())
	require.Equal(t, int16(4), keyObj.GetInt16())
	require.Equal(t, int32(4), keyObj.GetInt32())
	require.Equal(t, int64(4), keyObj.GetInt64())
	require.Equal(t, uint(4), keyObj.GetUint())
	require.Equal(t, uint8(4), keyObj.GetUint8())
	require.Equal(t, uint16(4), keyObj.GetUint16())
	require.Equal(t, uint32(4), keyObj.GetUint32())
	require.Equal(t, uint64(4), keyObj.GetUint64())
	require.Equal(t, float32(4), keyObj.GetFloat32())
	require.Equal(t, float64(4), keyObj.GetFloat64())
}

func TestParseJSON(t *testing.T) {
	testGroupName := "g5"
	testKeyName := "k5"

	value := `{"a": 1, "b": {"c": ["x", "y", "z"]}}`
	err := testProvider.Set(testGroupName, testKeyName, []byte(value))
	require.Nil(t, err)

	keyObj, err := newWatchKeyObject(testGroupName, testKeyName)
	require.Nil(t, err)

	require.Equal(t, keyObj.GetString(), value)

	var a struct {
		A int `json:"a"`
		B struct {
			C []string `json:"c"`
		} `json:"b"`
	}
	err = keyObj.ParseJSON(&a)
	require.Nil(t, err)
	require.Equal(t, 1, a.A)
	require.Equal(t, []string{"x", "y", "z"}, a.B.C)
}

func TestParseYaml(t *testing.T) {
	testGroupName := "g6"
	testKeyName := "k6"

	value := `
a: 1
b:
  c:
    - x
    - y
    - z`
	err := testProvider.Set(testGroupName, testKeyName, []byte(value))
	require.Nil(t, err)

	keyObj, err := newWatchKeyObject(testGroupName, testKeyName)
	require.Nil(t, err)

	require.Equal(t, keyObj.GetString(), value)

	var a struct {
		A int `json:"a"`
		B struct {
			C []string `json:"c"`
		} `json:"b"`
	}
	err = keyObj.ParseYaml(&a)
	require.Nil(t, err)
	require.Equal(t, 1, a.A)
	require.Equal(t, []string{"x", "y", "z"}, a.B.C)
}