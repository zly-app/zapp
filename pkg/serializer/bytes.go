package serializer

import (
	"fmt"
	"io"
	"reflect"
	"unsafe"
)

const BytesSerializerName = "bytes"

type bytesSerializer struct{}

func (bytesSerializer) toBytes(a interface{}) ([]byte, error) {
	switch v := a.(type) {
	case []byte:
		return v, nil
	case *[]byte:
		return *v, nil
	case string:
		return StringToBytes(&v), nil
	case *string:
		return StringToBytes(v), nil
	}
	return nil, fmt.Errorf("a not bytes, it's %T", a)
}

func (b bytesSerializer) Marshal(a interface{}, w io.Writer) error {
	bs, err := b.toBytes(a)
	if err != nil {
		return err
	}
	_, err = w.Write(bs)
	return err
}

func (b bytesSerializer) MarshalBytes(a interface{}) ([]byte, error) {
	bs, err := b.toBytes(a)
	if err != nil {
		return nil, err
	}
	ret := make([]byte, len(bs))
	copy(ret, bs)
	return ret, nil
}

func (bytesSerializer) Unmarshal(r io.Reader, a interface{}) error {
	bs, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	switch v := a.(type) {
	case *[]byte:
		*v = bs
		return nil
	case *string:
		*v = *BytesToString(bs)
		return nil
	}
	return fmt.Errorf("a not bytes, it's %T", a)
}

func (bytesSerializer) UnmarshalBytes(data []byte, a interface{}) error {
	switch v := a.(type) {
	case *[]byte:
		ret := make([]byte, len(data))
		copy(ret, data)
		*v = ret
		return nil
	case *string:
		ret := make([]byte, len(data))
		copy(ret, data)
		*v = *BytesToString(ret)
		return nil
	}
	return fmt.Errorf("a not bytes, it's %T", a)
}

func StringToBytes(s *string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}
func BytesToString(b []byte) *string {
	return (*string)(unsafe.Pointer(&b))
}
