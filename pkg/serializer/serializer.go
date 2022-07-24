package serializer

import (
	"io"

	"go.uber.org/zap"

	"github.com/zly-app/zapp/logger"
)

// 序列化器
type ISerializer interface {
	// 序列化
	Marshal(a interface{}, w io.Writer) error
	// 反序列化
	Unmarshal(in io.Reader, a interface{}) error
}

var serializers = map[string]ISerializer{
	JsonSerializerName:             jsonSerializer{},
	JsonIterSerializerName:         jsonIterSerializer{},
	JsonIterStandardSerializerName: jsonIterStandardSerializer{},
	MsgPackSerializerName:          msgPackSerializer{},
	YamlSerializerName:             yamlSerializer{},
}

// 注册序列化器, 重复注册会panic
func RegistrySerializer(name string, c ISerializer, replace ...bool) {
	if len(replace) == 0 || !replace[0] {
		if _, ok := serializers[name]; ok {
			logger.Log.Panic("Serializer重复注册", zap.String("name", name))
		}
	}
	serializers[name] = c
}

// 获取序列化器
func GetSerializer(name string) ISerializer {
	c, ok := serializers[name]
	if !ok {
		logger.Log.Panic("试图获取未注册的序列化器", zap.String("name", name))
	}
	return c
}

// 尝试获取序列化器
func TryGetSerializer(name string) (ISerializer, bool) {
	c, ok := serializers[name]
	return c, ok
}
