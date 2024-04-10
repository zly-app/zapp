package core

// 配置观察选项
type ConfigWatchOption func(opts interface{})

// 配置观察key对象
type IConfigWatchKeyObject interface {
	// 获取组名
	GroupName() string
	// 获取key名
	KeyName() string
	// 添加回调, 即使没有发生变更, 启动时会立即触发一次回调
	AddCallback(callback ...ConfigWatchKeyCallback)
	// 获取原始数据的副本
	GetData() []byte
	// 检查是否复合预期的值
	Expect(v interface{}) bool
	// 获取字符串
	GetString() string
	GetBool(def ...bool) bool
	GetInt(def ...int) int
	GetInt8(def ...int8) int8
	GetInt16(def ...int16) int16
	GetInt32(def ...int32) int32
	GetInt64(def ...int64) int64
	GetUint(def ...uint) uint
	GetUint8(def ...uint8) uint8
	GetUint16(def ...uint16) uint16
	GetUint32(def ...uint32) uint32
	GetUint64(def ...uint64) uint64
	GetFloat32(def ...float32) float32
	GetFloat64(def ...float64) float64

	/*解析为json
	  outPtr 用于接收数据的指针
	*/
	ParseJSON(outPtr interface{}) error
	/*解析为yaml
	  outPtr 用于接收数据的指针
	*/
	ParseYaml(outPtr interface{}) error
}

// 配置观察key对象回调, 如果是第一次触发, isInit 为 true
type ConfigWatchKeyCallback func(isInit bool, oldData, newData []byte)

// 配置观察key对象, 用于结构化
type IConfigWatchKeyStruct[T any] interface {
	// 获取组名
	GroupName() string
	// 获取key名
	KeyName() string
	// 添加回调, 即使没有发生变更, 启动时也会触发一次回调
	AddCallback(callback ...ConfigWatchKeyStructCallback[T])
	// 获取原始数据的副本
	GetData() []byte
	// 获取结构
	Get() T
}

// 配置观察key对象回调, 如果是第一次触发, isInit 为 true
type ConfigWatchKeyStructCallback[T any] func(isInit bool, oldData, newData T)

// 配置观察提供者
type IConfigWatchProvider interface {
	// 获取数据
	Get(groupName, keyName string) ([]byte, error)
	// 监听, 注意, 这个方法不能一直阻塞, 应该尽早的返回, 而通过协程开始watch
	Watch(groupName, keyName string, callback ConfigWatchProviderCallback) error
}

// 配置观察提供者回调
type ConfigWatchProviderCallback func(groupName, keyName string, oldData, newData []byte)
