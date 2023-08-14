package serializer

import (
	"io"

	"github.com/bytedance/sonic"
)

const SonicSerializerName = "sonic"

// 使用第三方包sonic进行序列化
type sonicSerializer struct{}

var sonicDefault = sonic.ConfigDefault

func (sonicSerializer) Marshal(a interface{}, w io.Writer) error {
	return sonicDefault.NewEncoder(w).Encode(a)
}

func (s sonicSerializer) MarshalBytes(a interface{}) ([]byte, error) {
	return sonicDefault.Marshal(a)
}

func (sonicSerializer) Unmarshal(r io.Reader, a interface{}) error {
	return sonicDefault.NewDecoder(r).Decode(a)
}

func (s sonicSerializer) UnmarshalBytes(data []byte, a interface{}) error {
	return sonicDefault.Unmarshal(data, a)
}
