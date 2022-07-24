package serializer

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

// jsonIter 实现的模拟内置 json 库
const JsonIterStandardSerializerName = "jsoniter_standard"

type jsonIterStandardSerializer struct{}

func (jsonIterStandardSerializer) Marshal(a interface{}, w io.Writer) error {
	return jsoniter.ConfigCompatibleWithStandardLibrary.NewEncoder(w).Encode(a)
}

func (jsonIterStandardSerializer) Unmarshal(in io.Reader, a interface{}) error {
	return jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder(in).Decode(a)
}
