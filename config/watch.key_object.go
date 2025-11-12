package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/log"
)

// 观察选项
type watchKeyObject struct {
	p    core.IConfigWatchProvider
	opts *watchOptions

	groupName string
	keyName   string
	initOpts  []core.ConfigWatchOption

	callbacks []core.ConfigWatchKeyCallback
	watchMx   sync.Mutex // 用于锁 callback

	data   atomic.Value
	initWG sync.WaitGroup
}

func (w *watchKeyObject) init() {
	w.initWG.Add(1)
	go func() {
		waitAppInit()
		w.opts = newWatchOptions(w.initOpts)
		w.p = w.opts.Provider

		// 立即获取
		data, err := w.p.Get(w.groupName, w.keyName)
		if err != nil {
			log.Log.Fatal("获取配置失败",
				zap.String("groupName", w.groupName),
				zap.String("keyName", w.keyName),
				zap.Error(err))
		}
		w.resetData(data)

		// 开始观察
		err = w.p.Watch(w.groupName, w.keyName, w.watchCallback)
		if err != nil {
			log.Log.Fatal("watch配置失败",
				zap.String("groupName", w.groupName),
				zap.String("keyName", w.keyName),
				zap.Error(err))
		}

		w.initWG.Done()
	}()
}

// 等待初始化
func (w *watchKeyObject) waitInit() {
	w.initWG.Wait()
}

func (w *watchKeyObject) Opts() *watchOptions {
	return w.opts
}

func (w *watchKeyObject) GroupName() string { return w.groupName }
func (w *watchKeyObject) KeyName() string   { return w.keyName }

func (w *watchKeyObject) AddCallback(callback ...core.ConfigWatchKeyCallback) {
	w.waitInit()
	w.watchMx.Lock()
	defer w.watchMx.Unlock()

	items := make([]core.ConfigWatchKeyCallback, 0, len(callback))
	items = append(items, callback...)
	w.callbacks = append(w.callbacks, items...)

	// 立即触发
	data := w.getRawData()
	for _, fn := range callback {
		fn(true, data, data) // 这里无法保证 data 被 callback 函数修改数据
	}
}

func (w *watchKeyObject) GetData() []byte {
	w.waitInit()
	return w.getRawData()
}

// 检查是否复合预期的值
func (w *watchKeyObject) Expect(v interface{}) bool {
	w.waitInit()
	switch t := v.(type) {
	case []byte:
		return bytes.Equal(t, w.getRawData())
	case string:
		return bytes.Equal([]byte(t), w.getRawData())
	case []rune:
		return bytes.Equal([]byte(string(t)), w.getRawData())
	case bool:
		temp, err := w.getBool()
		if err != nil {
			return false
		}
		return temp == t
	case int:
		return w.GetInt(t+1) == t
	case int8:
		return w.GetInt8(t+1) == t
	case int16:
		return w.GetInt16(t+1) == t
	case int32:
		return w.GetInt32(t+1) == t
	case int64:
		return w.GetInt64(t+1) == t
	case uint:
		return w.GetUint(t+1) == t
	case uint8:
		return w.GetUint8(t+1) == t
	case uint16:
		return w.GetUint16(t+1) == t
	case uint32:
		return w.GetUint32(t+1) == t
	case uint64:
		return w.GetUint64(t+1) == t
	case float32:
		return w.GetFloat32(t+1) == t
	case float64:
		return w.GetFloat64(t+1) == t
	}
	return false
}

func (w *watchKeyObject) GetString() string {
	w.waitInit()
	return string(w.getRawData())
}
func (w *watchKeyObject) getBool() (bool, error) {
	switch v := w.GetString(); v {
	case "1", "t", "T", "true", "TRUE", "True", "y", "Y", "yes", "YES", "Yes",
		"on", "ON", "On", "ok", "OK", "Ok",
		"enabled", "ENABLED", "Enabled",
		"open", "OPEN", "Open":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False", "n", "N", "no", "NO", "No",
		"off", "OFF", "Off", "cancel", "CANCEL", "Cancel",
		"disable", "DISABLE", "Disable",
		"close", "CLOSE", "Close":
		return false, nil
	default:
		return false, fmt.Errorf("data %s can't conver to boolean", v)
	}
}
func (w *watchKeyObject) GetBool(def ...bool) bool {
	w.waitInit()
	v, err := w.getBool()
	if err == nil {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return false
}
func (w *watchKeyObject) GetInt(def ...int) int {
	w.waitInit()
	v, err := strconv.Atoi(w.GetString())
	if err == nil {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (w *watchKeyObject) GetInt8(def ...int8) int8 {
	w.waitInit()
	v, err := strconv.ParseInt(w.GetString(), 10, 8)
	if err == nil {
		return int8(v)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (w *watchKeyObject) GetInt16(def ...int16) int16 {
	w.waitInit()
	v, err := strconv.ParseInt(w.GetString(), 10, 16)
	if err == nil {
		return int16(v)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (w *watchKeyObject) GetInt32(def ...int32) int32 {
	w.waitInit()
	v, err := strconv.ParseInt(w.GetString(), 10, 32)
	if err == nil {
		return int32(v)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (w *watchKeyObject) GetInt64(def ...int64) int64 {
	w.waitInit()
	v, err := strconv.ParseInt(w.GetString(), 10, 64)
	if err == nil {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (w *watchKeyObject) GetUint(def ...uint) uint {
	w.waitInit()
	v, err := strconv.ParseUint(w.GetString(), 10, 64)
	if err == nil {
		return uint(v)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (w *watchKeyObject) GetUint8(def ...uint8) uint8 {
	w.waitInit()
	v, err := strconv.ParseUint(w.GetString(), 10, 8)
	if err == nil {
		return uint8(v)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (w *watchKeyObject) GetUint16(def ...uint16) uint16 {
	w.waitInit()
	v, err := strconv.ParseUint(w.GetString(), 10, 16)
	if err == nil {
		return uint16(v)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (w *watchKeyObject) GetUint32(def ...uint32) uint32 {
	w.waitInit()
	v, err := strconv.ParseUint(w.GetString(), 10, 32)
	if err == nil {
		return uint32(v)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (w *watchKeyObject) GetUint64(def ...uint64) uint64 {
	w.waitInit()
	v, err := strconv.ParseUint(w.GetString(), 10, 64)
	if err == nil {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (w *watchKeyObject) GetFloat32(def ...float32) float32 {
	w.waitInit()
	v, err := strconv.ParseFloat(w.GetString(), 32)
	if err == nil {
		return float32(v)
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
func (w *watchKeyObject) GetFloat64(def ...float64) float64 {
	w.waitInit()
	v, err := strconv.ParseFloat(w.GetString(), 64)
	if err == nil {
		return v
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

func (w *watchKeyObject) ParseJSON(outPtr interface{}) error {
	w.waitInit()
	return json.Unmarshal(w.getRawData(), outPtr)
}
func (w *watchKeyObject) ParseYaml(outPtr interface{}) error {
	w.waitInit()
	return yaml.Unmarshal(w.getRawData(), outPtr)
}

// 获取原始数据
func (w *watchKeyObject) getRawData() []byte {
	data := w.data.Load().([]byte)
	return data
}

// 回调
func (w *watchKeyObject) watchCallback(_, _ string, _, newData []byte) {
	oldData := w.getRawData()
	if bytes.Equal(newData, oldData) {
		return
	}

	w.resetData(newData)
	log.Log.Info("配置数据变更",
		zap.String("groupName", w.groupName),
		zap.String("keyName", w.keyName),
		zap.String("oldData", string(oldData)),
		zap.String("newData", string(newData)),
	)

	w.watchMx.Lock()
	defer w.watchMx.Unlock()
	for _, fn := range w.callbacks {
		go fn(false, oldData, newData) // 这里无法保证 newData 被 callback 函数修改数据
	}
}

// 重新设置数据
func (w *watchKeyObject) resetData(data []byte) {
	if data == nil {
		data = make([]byte, 0)
	}
	w.data.Store(data)
}

func newWatchKeyObject(groupName, keyName string, opts ...core.ConfigWatchOption) *watchKeyObject {
	w := &watchKeyObject{
		groupName: groupName,
		keyName:   keyName,
		initOpts:  opts,
	}
	w.init()
	return w
}

// 观察key, 失败会fatal
func WatchKey(groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyObject {
	w := newWatchKeyObject(groupName, keyName, opts...)
	return w
}
