package serializer

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

const JsonIterSerializerName = "jsoniter"

// 使用第三方包json-iterator进行序列化
type jsonIterSerializer struct{}

func (jsonIterSerializer) Marshal(a interface{}, w io.Writer) error {
	return jsoniter.NewEncoder(w).Encode(a)
}

func (jsonIterSerializer) Unmarshal(in io.Reader, a interface{}) error {
	return jsoniter.NewDecoder(in).Decode(a)
}
