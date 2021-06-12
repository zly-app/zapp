/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/4/17
   Description :
-------------------------------------------------
*/

package zlog

import (
	"fmt"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zly-app/zapp/core"
)

var loggerId uint32

// 获取下个日志id
//
// 将数值转为32进制, 因为求余2的次幂可以用位运算所以采用 数字+22位英文字母
func (l *logWrap) nextLoggerId() string {
	id := atomic.AddUint32(&loggerId, 1)
	bs := []byte{48, 48, 48, 48, 48, 48, 48}

	i := len(bs) - 1
	for {
		bs[i] = byte(31&id) + 48 // 从字符0开始
		if bs[i] > 57 {          // 超过数字用字母表示
			bs[i] += 39
		}
		if id < 32 {
			return string(bs)
		}
		i--
		id >>= 5
	}
}

type logWrap struct {
	log            *zap.Logger
	fields         []zap.Field
	callerMinLevel zapcore.Level
	ws             zapcore.WriteSyncer
}

var _ core.ILogger = (*logWrap)(nil)

func newLogWrap(log *zap.Logger, callerMinLevel zapcore.Level, ws zapcore.WriteSyncer) *logWrap {
	l := &logWrap{
		log:            log,
		fields:         nil,
		callerMinLevel: callerMinLevel,
		ws:             ws,
	}
	return l
}

func (l *logWrap) print(level Level, v []interface{}) {
	msg, fields := l.makeBody(v)
	zapLevel := parserLogLevel(level)
	if ce := l.log.Check(zapLevel, msg); ce != nil {
		if zapLevel < l.callerMinLevel {
			ce.Caller.Defined = false
		}
		ce.Write(fields...)
	}
}

func (l *logWrap) makeBody(v []interface{}) (string, []zap.Field) {
	args := make([]interface{}, 0, len(v))
	fields := append([]zap.Field{}, l.fields...)
	for _, value := range v {
		switch val := value.(type) {
		case zap.Field:
			fields = append(fields, val)
		case *zap.Field:
			fields = append(fields, *val)
		default:
			args = append(args, value)
		}
	}
	return fmt.Sprint(args...), fields
}

func (l *logWrap) Log(level Level, v ...interface{}) {
	l.print(level, v)
}
func (l *logWrap) Debug(v ...interface{}) {
	l.print(DebugLevel, v)
}
func (l *logWrap) Info(v ...interface{}) {
	l.print(InfoLevel, v)
}
func (l *logWrap) Warn(v ...interface{}) {
	l.print(WarnLevel, v)
}
func (l *logWrap) Error(v ...interface{}) {
	l.print(ErrorLevel, v)
}
func (l *logWrap) DPanic(v ...interface{}) {
	l.print(DPanicLevel, v)
}
func (l *logWrap) Panic(v ...interface{}) {
	l.print(PanicLevel, v)
}
func (l *logWrap) Fatal(v ...interface{}) {
	l.print(FatalLevel, v)
}

// 创建一个会话log
func (l *logWrap) NewSessionLogger(fields ...zap.Field) core.ILogger {
	return &logWrap{
		log:            l.log,
		fields:         append(append(append([]zap.Field{}, l.fields...), zap.String(logIdKey, l.nextLoggerId())), fields...),
		callerMinLevel: l.callerMinLevel,
		ws:             l.ws,
	}
}

// 获取日志输出合成器
func GetLogWriteSyncer(l interface{}) (zapcore.WriteSyncer, bool) {
	if a, ok := l.(*logWrap); ok {
		return a.ws, true
	}
	return nil, false
}
