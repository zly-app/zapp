package serializer

import (
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

const MsgPackSerializerName = "msgpack"

// MsgPack序列化器
type msgPackSerializer struct{}

func (msgPackSerializer) Marshal(a interface{}, w io.Writer) error {
	enc := msgpack.NewEncoder(w)
	enc.SetCustomStructTag("json") // 如果没有 msgpack 标记, 使用 json 标记
	return enc.Encode(a)
}

func (msgPackSerializer) Unmarshal(in io.Reader, a interface{}) error {
	dec := msgpack.NewDecoder(in)
	dec.SetCustomStructTag("json") // 如果没有 msgpack 标记, 使用 json 标记
	err := dec.Decode(a)
	return err
}
