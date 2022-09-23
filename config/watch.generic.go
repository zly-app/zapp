package config

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

// 观察选项
type watchKeyGeneric[T any] struct {
	w *watchKeyObject

	callbacks []core.ConfigWatchKeyStructCallback[T] // 必须自己管理, 因为 watchKeyObject 是通过协程调用的 callback
	watchMx   sync.Mutex                             // 用于锁 callback

	rawData atomic.Value // 这里只保留成功解析的数据
	data    atomic.Value
}

func (w *watchKeyGeneric[T]) GroupName() string { return w.w.GroupName() }
func (w *watchKeyGeneric[T]) KeyName() string   { return w.w.KeyName() }

func (w *watchKeyGeneric[T]) AddCallback(callback ...core.ConfigWatchKeyStructCallback[T]) {
	w.watchMx.Lock()
	defer w.watchMx.Unlock()

	items := make([]core.ConfigWatchKeyStructCallback[T], 0, len(callback))
	items = append(items, callback...)
	w.callbacks = append(w.callbacks, items...)

	// 立即触发
	data := w.Get()
	for _, fn := range callback {
		fn(w, true, data, data) // 这里无法保证 data 被 callback 函数修改数据
	}
}

func (w *watchKeyGeneric[T]) GetData() []byte {
	data := w.rawData.Load().([]byte)
	return data
}

func (w *watchKeyGeneric[T]) Get() T {
	data := w.data.Load().(*T)
	return *data
}

// 重新解析数据
func (w *watchKeyGeneric[T]) reset(first bool, newData []byte) error {
	data := new(T)
	var err error
	switch t := w.w.Opts().StructType; t {
	case Json:
		err = w.w.ParseJSON(data)
	case Yaml:
		err = w.w.ParseYaml(data)
	default:
		err = fmt.Errorf("未定义的解析类型: %v", t)
	}
	if err != nil {
		return fmt.Errorf("解析配置失败: %v", err)
	}

	var oldData T
	if !first {
		oldData = w.Get()
	}
	w.data.Store(data)
	w.rawData.Store(newData)

	if first {
		return nil
	}

	w.watchMx.Lock()
	defer w.watchMx.Unlock()
	for _, fn := range w.callbacks {
		go fn(w, false, oldData, *data) // 这里无法保证 newData 被 callback 函数修改数据
	}
	return nil
}

func (w *watchKeyGeneric[T]) watchCallback(_ core.IConfigWatchKeyObject, first bool, _, newData []byte) {
	err := w.reset(first, newData)
	if err == nil {
		return
	}
	if first {
		logger.Log.Fatal("重置数据失败",
			zap.String("groupName", w.GroupName()),
			zap.String("keyName", w.KeyName()),
			zap.String("newData", string(newData)),
			zap.Error(err),
		)
	}
	logger.Log.Error("重置数据失败",
		zap.String("groupName", w.GroupName()),
		zap.String("keyName", w.KeyName()),
		zap.String("newData", string(newData)),
		zap.Error(err),
	)
}

func newWatchKeyStruct[T any](groupName, keyName string, opts ...core.ConfigWatchOption) (
	core.IConfigWatchKeyStruct[T], error) {
	temp := new(T)
	vTemp := reflect.ValueOf(temp).Elem() // 消除new指针
	if vTemp.Kind() == reflect.Ptr {
		return nil, fmt.Errorf("泛型类型不能是指针, T=%T", *temp)
	}

	w, err := newWatchKeyObject(groupName, keyName, opts...)
	if err != nil {
		return nil, fmt.Errorf("观察key失败: %v", err)
	}
	warp := &watchKeyGeneric[T]{
		w: w,
	}
	// 这里会立即触发, 所以下一步可以立即返回 warp
	w.AddCallback(warp.watchCallback)
	return warp, nil
}

// 观察key结构化数据, 失败会fatal
func WatchKeyStruct[T any](groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyStruct[T] {
	w, err := newWatchKeyStruct[T](groupName, keyName, opts...)
	if err != nil {
		logger.Log.Fatal("观察key失败",
			zap.String("groupName", groupName),
			zap.String("keyName", keyName),
			zap.Error(err),
		)
	}
	return w
}
