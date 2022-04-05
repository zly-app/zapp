package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

// 观察选项
type watchKeyObject struct {
	p core.IConfigWatchProvider

	groupName string
	keyName   string
	callbacks []core.IConfigWatchKeyCallback

	data atomic.Value
}

func (w *watchKeyObject) GroupName() string { return w.groupName }
func (w *watchKeyObject) KeyName() string   { return w.keyName }

func (w *watchKeyObject) AddCallback(callback ...core.IConfigWatchKeyCallback) {
	items := make([]core.IConfigWatchKeyCallback, 0, len(callback))
	items = append(items, callback...)
	w.callbacks = append(w.callbacks, items...)
}

func (w *watchKeyObject) GetData() []byte {
	data := w.getRawData()
	copyData := make([]byte, len(data))
	copy(copyData, data)
	return copyData
}

// 检查是否复合预期的值
func (w *watchKeyObject) Expect(v interface{}) bool {
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

func (w *watchKeyObject) GetString() string { return string(w.getRawData()) }
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
	return json.Unmarshal(w.getRawData(), outPtr)
}
func (w *watchKeyObject) ParseYaml(outPtr interface{}) error {
	return yaml.Unmarshal(w.getRawData(), outPtr)
}

// 检查状态
func (w *watchKeyObject) check() {
	if w.p == nil {
		w.p = GetDefaultConfigWatchProvider()
	}
	if w.p == nil {
		logger.Log.Fatal("默认配置观察提供者不存在")
	}
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

	data := make([]byte, len(newData))
	copy(data, newData)
	w.resetData(data)

	for _, fn := range w.callbacks {
		fn(w, oldData, w.GetData()) // 通过GetData重新获取数据保证不会被改变
	}
}

// 重新设置数据
func (w *watchKeyObject) resetData(data []byte) {
	if data == nil {
		data = make([]byte, 0)
	}
	w.data.Store(data)
}

func newWatchKeyObject(groupName, keyName string, opts ...core.ConfigWatchOption) core.IConfigWatchKeyObject {
	w := &watchKeyObject{
		groupName: groupName,
		keyName:   keyName,
	}
	for _, opt := range opts {
		opt(w)
	}

	w.check()
	w.resetData(nil)

	// 立即获取
	data, err := w.p.Get(groupName, keyName)
	if err != nil {
		logger.Log.Fatal("获取watch数据失败",
			zap.String("groupName", groupName),
			zap.String("keyName", keyName),
			zap.Error(err),
		)
	}

	// 立即触发
	w.watchCallback(groupName, keyName, nil, data)

	// 开始观察
	err = w.p.Watch(groupName, keyName, w.watchCallback)
	if err != nil {
		logger.Log.Fatal("watch配置失败",
			zap.String("groupName", groupName),
			zap.String("keyName", keyName),
			zap.Error(err),
		)
	}
	return w
}
