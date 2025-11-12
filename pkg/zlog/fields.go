package zlog

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func convertFieldValue(f zap.Field) string {
	switch f.Type {
	case zapcore.StringType:
		return f.String
	case zapcore.BoolType:
		if f.Integer == 1 {
			return "true"
		}
		return "false"
	case zapcore.Float64Type, zapcore.Float32Type,
		zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type,
		zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type,
		zapcore.UintptrType:
		return strconv.FormatInt(f.Integer, 10)
	case zapcore.TimeType:
		t := time.Unix(0, f.Integer).In(time.Local)
		return t.Format(time.RFC3339)
	case zapcore.TimeFullType:
		t, ok := f.Interface.(time.Time)
		if ok {
			return t.Format(time.RFC3339)
		}
		return "0"
	case zapcore.DurationType:
		return time.Duration(f.Integer).String()
	case zapcore.BinaryType:
		b, ok := f.Interface.([]byte)
		if ok {
			return base64.StdEncoding.EncodeToString(b)
		}
		return "binary data"
	case zapcore.Complex64Type, zapcore.Complex128Type:
		return fmt.Sprintf("%v", f.Interface)
	case zapcore.ByteStringType:
		b, ok := f.Interface.([]byte)
		if ok {
			return string(b)
		}
	case zapcore.StringerType:
		s, ok := f.Interface.(fmt.Stringer)
		if ok {
			return s.String()
		}
	}

	if f.String != "" {
		return f.String
	}
	if f.Integer != 0 {
		return strconv.FormatInt(f.Integer, 10)
	}
	if f.Interface == nil {
		return "nil"
	}
	text, _ := sonic.MarshalString(f.Interface)
	return text
}

type withoutAttachLog2Trace struct{}

// 不要将日志附加给trace
func WithoutAttachLog2Trace() interface{} {
	return withoutAttachLog2Trace{}
}
