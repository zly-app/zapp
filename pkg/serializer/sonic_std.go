package serializer

import (
	"io"

	"github.com/bytedance/sonic"
)

const SonicStdSerializerName = "sonic_std"

// 使用第三方包sonic进行序列化
type sonicStdSerializer struct{}

var sonicStd = sonic.ConfigStd

func (sonicStdSerializer) Marshal(a interface{}, w io.Writer) error {
	return sonicStd.NewEncoder(w).Encode(a)
}

func (s sonicStdSerializer) MarshalBytes(a interface{}) ([]byte, error) {
	return sonicStd.Marshal(a)
}

func (sonicStdSerializer) Unmarshal(r io.Reader, a interface{}) error {
	return sonicStd.NewDecoder(r).Decode(a)
}

func (s sonicStdSerializer) UnmarshalBytes(data []byte, a interface{}) error {
	return sonicStd.Unmarshal(data, a)
}
