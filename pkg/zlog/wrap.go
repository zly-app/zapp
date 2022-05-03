/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/4/17
   Description :
-------------------------------------------------
*/

package zlog

import (
	"context"
	"fmt"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
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

// 添加fields
func (l *logWrap) AddFields(fields ...zap.Field) {
	l.fields = append(append([]zap.Field{}, l.fields...), fields...)
}

/*为log移除一些field, 返回移除的个数
  count	移除key的个数, <1表示所有
*/
func (l *logWrap) RemoveFields(count int, fieldKeys ...string) int {
	if len(l.fields) == 0 || len(fieldKeys) == 0 {
		return 0
	}

	// key查找
	hasKey := func() func(key string) bool {
		// 值少时通过遍历方式
		if len(fieldKeys) <= 16 {
			return func(key string) bool {
				for _, k := range fieldKeys {
					if k == key {
						return true
					}
				}
				return false
			}
		}

		// 通过map方式
		keyMap := make(map[string]struct{}, len(fieldKeys))
		for _, k := range fieldKeys {
			keyMap[k] = struct{}{}
		}
		return func(key string) bool {
			_, ok := keyMap[key]
			return ok
		}
	}()

	var n int
	ff := make([]zap.Field, 0, len(l.fields))
	for i, f := range l.fields {
		if !hasKey(f.Key) {
			ff = append(ff, f)
			continue
		}

		n++
		if count >= 1 && n == count { // 不能再去掉了
			ff = append(ff, l.fields[i+1:]...) // 接上末尾的fields
			break
		}
	}
	l.fields = ff
	return n
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

// 创建一个带链路id的log
func (l *logWrap) NewTraceLogger(ctx context.Context, fields ...zap.Field) core.ILogger {
	span := utils.Trace.GetSpan(ctx)
	traceID := utils.Trace.GetTraceID(span)
	return &logWrap{
		log:            l.log,
		fields:         append(append(append([]zap.Field{}, l.fields...), zap.String(logTraceIdKey, traceID)), fields...),
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

// 为log添加一些field
func AddFields(l interface{}, fields ...zap.Field) bool {
	a, ok := l.(*logWrap)
	if !ok {
		return false
	}
	a.AddFields(fields...)
	return true
}

/*为log移除一些field, 返回移除的个数
  count	移除key的个数, <1表示所有
*/
func RemoveFields(l interface{}, count int, fieldKeys ...string) (int, bool) {
	a, ok := l.(*logWrap)
	if !ok {
		return 0, false
	}
	return a.RemoveFields(count, fieldKeys...), true
}
