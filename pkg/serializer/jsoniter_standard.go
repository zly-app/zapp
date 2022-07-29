package serializer

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

// jsonIter 实现的模拟内置 json 库
const JsonIterStandardSerializerName = "jsoniter_standard"

type jsonIterStandardSerializer struct{}

var jsonIterStandard = jsoniter.ConfigCompatibleWithStandardLibrary

func (jsonIterStandardSerializer) Marshal(a interface{}, w io.Writer) error {
	return jsonIterStandard.NewEncoder(w).Encode(a)
}

func (s jsonIterStandardSerializer) MarshalBytes(a interface{}) ([]byte, error) {
	return jsonIterStandard.Marshal(a)
}

func (jsonIterStandardSerializer) Unmarshal(r io.Reader, a interface{}) error {
	return jsonIterStandard.NewDecoder(r).Decode(a)
}

func (s jsonIterStandardSerializer) UnmarshalBytes(data []byte, a interface{}) error {
	return jsonIterStandard.Unmarshal(data, a)
}
