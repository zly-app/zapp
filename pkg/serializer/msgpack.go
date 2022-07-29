package serializer

import (
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

const MsgPackSerializerName = "msgpack"

// MsgPack序列化器
type msgPackSerializer struct{}

func (msgPackSerializer) Marshal(a interface{}, w io.Writer) error {
	enc := msgpack.GetEncoder()
	enc.SetCustomStructTag("json") // 如果没有 msgpack 标记, 使用 json 标记
	enc.Reset(w)
	err := enc.Encode(a)
	msgpack.PutEncoder(enc)
	return err
}

func (s msgPackSerializer) MarshalBytes(a interface{}) ([]byte, error) {
	return msgpack.Marshal(a)
}

func (msgPackSerializer) Unmarshal(r io.Reader, a interface{}) error {
	dec := msgpack.GetDecoder()
	dec.SetCustomStructTag("json") // 如果没有 msgpack 标记, 使用 json 标记
	dec.Reset(r)
	err := dec.Decode(a)
	msgpack.PutDecoder(dec)
	return err
}

func (s msgPackSerializer) UnmarshalBytes(data []byte, a interface{}) error {
	return msgpack.Unmarshal(data, a)
}
