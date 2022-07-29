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

func (s jsonIterSerializer) MarshalBytes(a interface{}) ([]byte, error) {
	return jsoniter.Marshal(a)
}

func (jsonIterSerializer) Unmarshal(r io.Reader, a interface{}) error {
	return jsoniter.NewDecoder(r).Decode(a)
}

func (s jsonIterSerializer) UnmarshalBytes(data []byte, a interface{}) error {
	return jsoniter.Unmarshal(data, a)
}
