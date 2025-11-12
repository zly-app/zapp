package config

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/log"
)

// 观察选项
type watchKeyGeneric[T any] struct {
	keyObject *watchKeyObject

	callbacks []core.ConfigWatchKeyStructCallback[T] // 必须自己管理, 因为 watchKeyObject 是通过协程调用的 callback
	watchMx   sync.Mutex                             // 用于锁 callback

	rawData  atomic.Value // 这里只保留成功解析的数据
	data     atomic.Value
	dataType reflect.Type

	initWG sync.WaitGroup
}

func (w *watchKeyGeneric[T]) GroupName() string { return w.keyObject.GroupName() }
func (w *watchKeyGeneric[T]) KeyName() string   { return w.keyObject.KeyName() }

func (w *watchKeyGeneric[T]) AddCallback(callback ...core.ConfigWatchKeyStructCallback[T]) {
	w.waitInit()
	w.watchMx.Lock()
	defer w.watchMx.Unlock()

	items := make([]core.ConfigWatchKeyStructCallback[T], 0, len(callback))
	items = append(items, callback...)
	w.callbacks = append(w.callbacks, items...)

	// 立即触发
	data := w.Get()
	for _, fn := range callback {
		fn(true, data, data) // 这里无法保证 data 被 callback 函数修改数据
	}
}

func (w *watchKeyGeneric[T]) GetData() []byte {
	w.waitInit()
	data := w.rawData.Load().([]byte)
	return data
}

func (w *watchKeyGeneric[T]) Get() T {
	w.waitInit()
	data := w.data.Load().(T)
	return data
}

// 重新解析数据
func (w *watchKeyGeneric[T]) reset(first bool, newData []byte) error {
	replica := reflect.New(w.dataType).Interface()
	var err error
	switch t := w.keyObject.Opts().StructType; t {
	case Json:
		err = w.keyObject.ParseJSON(replica)
	case Yaml:
		err = w.keyObject.ParseYaml(replica)
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

	// 最终注入的hook
	resultData := newData
	var cancel bool
	for _, hook := range resetInjectStructuredHooks {
		if replica, resultData, cancel = hook(replica, resultData); cancel {
			log.Log.Warn(fmt.Sprintf(
				"注入对象已被解析, 但是被拦截了! group: %s, key: %s, obj: %+v",
				w.GroupName(), w.KeyName(), replica))
			return nil
		}
	}

	w.data.Store(replica)
	w.rawData.Store(resultData)

	w.watchMx.Lock()
	defer w.watchMx.Unlock()
	for _, fn := range w.callbacks {
		go fn(first, oldData, replica.(T)) // 这里无法保证 newData 被 callback 函数修改数据
	}
	return nil
}

func (w *watchKeyGeneric[T]) watchCallback(first bool, oldData, newData []byte) {
	err := w.reset(first, newData)
	if err == nil {
		return
	}
	if first {
		log.Log.Fatal("首次解析数据失败",
			zap.String("groupName", w.GroupName()),
			zap.String("keyName", w.KeyName()),
			zap.String("data", string(newData)),
			zap.Error(err),
		)
	}
	log.Log.Error("重置数据失败",
		zap.String("groupName", w.GroupName()),
		zap.String("keyName", w.KeyName()),
		zap.String("oldData", string(oldData)),
		zap.String("newData", string(newData)),
		zap.Error(err),
	)
}

func (w *watchKeyGeneric[T]) init() {
	w.initWG.Add(1)
	go func() {
		w.keyObject.AddCallback(w.watchCallback) // 等待app初始化完毕后, 这里底层的w会立即触发回调
		w.initWG.Done()
	}()
}

// 等待初始化
func (w *watchKeyGeneric[T]) waitInit() {
	w.initWG.Wait()
}

func newWatchKeyStruct[T any](groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyStruct[T] {
	temp := *new(T) // 消除new指针
	vTemp := reflect.TypeOf(temp)
	if vTemp.Kind() != reflect.Ptr {
		log.Log.Fatal("泛型类型必须是指针", zap.String("T", fmt.Sprintf("%T", temp)))
	}

	warp := &watchKeyGeneric[T]{
		keyObject: newWatchKeyObject(groupName, keyName, opts...),
		dataType:  vTemp.Elem(), // 实际结构
	}
	warp.init()
	return warp
}

// 观察key结构化数据, 失败会fatal, 默认为json格式
func WatchKeyStruct[T any](groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyStruct[T] {
	w := newWatchKeyStruct[T](groupName, keyName, opts...)
	return w
}

// 观察json配置数据, 失败会fatal
func WatchJson[T any](groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyStruct[T] {
	opts = append(make([]core.ConfigWatchOption, 0, len(opts)+1), opts...)
	opts = append(opts, WithWatchStructJson())
	return newWatchKeyStruct[T](groupName, keyName, opts...)
}

// 观察yaml配置数据, 失败会fatal
func WatchYaml[T any](groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyStruct[T] {
	opts = append(make([]core.ConfigWatchOption, 0, len(opts)+1), opts...)
	opts = append(opts, WithWatchStructYaml())
	return newWatchKeyStruct[T](groupName, keyName, opts...)
}
