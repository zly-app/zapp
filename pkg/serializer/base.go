package serializer

import (
	"fmt"
	"io"
	"strconv"
)

const BaseSerializerName = "base"

type baseSerializer struct {
	marshal   func(a interface{}) ([]byte, error)
	unmarshal func(data []byte, a interface{}) error
}

func NewBaseSerializer(
	marshal func(a interface{}) ([]byte, error),
	unmarshal func(data []byte, a interface{}) error,
) ISerializer {
	return baseSerializer{marshal, unmarshal}
}

func (b baseSerializer) Marshal(a interface{}, w io.Writer) error {
	s, err := b.toStr(a)
	if err != nil {
		return err
	}
	_, err = w.Write(StringToBytes(&s))
	return err
}

func (b baseSerializer) toBool(s string) (bool, error) {
	switch s {
	case "1", "t", "T", "true", "TRUE", "True", "y", "Y", "yes", "YES", "Yes",
		"on", "ON", "On", "ok", "OK", "Ok",
		"enabled", "ENABLED", "Enabled",
		"open", "OPEN", "Open":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False", "n", "N", "no", "NO", "No",
		"off", "OFF", "Off", "cancel", "CANCEL", "Cancel",
		"disable", "DISABLE", "Disable",
		"close", "CLOSE", "Close",
		"", "nil", "Nil", "NIL", "null", "Null", "NULL", "none", "None", "NONE":
		return false, nil
	}
	return false, fmt.Errorf(`数据"%s"无法转换为bool`, s)
}
func (b baseSerializer) toStr(a interface{}) (string, error) {
	switch v := a.(type) {

	case nil:
		return "", nil
	case string:
		return v, nil
	case []byte:
		return *BytesToString(v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil

	case int:
		return strconv.FormatInt(int64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(v, 10), nil

	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	}

	switch v := a.(type) {
	case *string:
		return *v, nil
	case *[]byte:
		return *BytesToString(*v), nil
	case *bool:
		if *v {
			return "true", nil
		}
		return "false", nil

	case *int:
		return strconv.FormatInt(int64(*v), 10), nil
	case *int8:
		return strconv.FormatInt(int64(*v), 10), nil
	case *int16:
		return strconv.FormatInt(int64(*v), 10), nil
	case *int32:
		return strconv.FormatInt(int64(*v), 10), nil
	case *int64:
		return strconv.FormatInt(*v, 10), nil

	case *uint:
		return strconv.FormatUint(uint64(*v), 10), nil
	case *uint8:
		return strconv.FormatUint(uint64(*v), 10), nil
	case *uint16:
		return strconv.FormatUint(uint64(*v), 10), nil
	case *uint32:
		return strconv.FormatUint(uint64(*v), 10), nil
	case *uint64:
		return strconv.FormatUint(*v, 10), nil
	}

	data, err := b.marshal(a)
	if err != nil {
		return "", err
	}
	return *BytesToString(data), nil
}

func (b baseSerializer) MarshalBytes(a interface{}) ([]byte, error) {
	s, err := b.toStr(a)
	if err != nil {
		return nil, err
	}
	return StringToBytes(&s), nil
}

func (b baseSerializer) Unmarshal(r io.Reader, a interface{}) error {
	bs, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return b.unmarshalBytes(bs, a)
}

func (b baseSerializer) UnmarshalBytes(data []byte, a interface{}) (err error) {
	return b.unmarshalBytes(data, a)
}
func (b baseSerializer) unmarshalBytes(data []byte, a interface{}) (err error) {
	s := *BytesToString(data)
	switch p := a.(type) {
	case nil:
		return nil
	case *string:
		*p = s
	case *[]byte:
		*p = data
	case *bool:
		*p, err = b.toBool(s)
	case *int:
		*p, err = strconv.Atoi(s)
	case *int8:
		var n int64
		n, err = strconv.ParseInt(s, 10, 8)
		*p = int8(n)
	case *int16:
		var n int64
		n, err = strconv.ParseInt(s, 10, 16)
		*p = int16(n)
	case *int32:
		var n int64
		n, err = strconv.ParseInt(s, 10, 32)
		*p = int32(n)
	case *int64:
		*p, err = strconv.ParseInt(s, 10, 64)

	case *uint:
		var n uint64
		n, err = strconv.ParseUint(s, 10, 64)
		*p = uint(n)
	case *uint8:
		var n uint64
		n, err = strconv.ParseUint(s, 10, 8)
		*p = uint8(n)
	case *uint16:
		var n uint64
		n, err = strconv.ParseUint(s, 10, 16)
		*p = uint16(n)
	case *uint32:
		var n uint64
		n, err = strconv.ParseUint(s, 10, 32)
		*p = uint32(n)
	case *uint64:
		*p, err = strconv.ParseUint(s, 10, 64)

	case *float32:
		var n float64
		n, err = strconv.ParseFloat(s, 32)
		*p = float32(n)
	case *float64:
		*p, err = strconv.ParseFloat(s, 64)

	default:
		return b.unmarshal(data, a)
	}
	return
}
