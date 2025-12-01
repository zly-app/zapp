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
	"strconv"
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
func (l *logCore) nextLoggerId() string {
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

type logCore struct {
	conf           *core.LogConfig
	log            *zap.Logger
	fields         []zap.Field
	level          zapcore.Level
	callerMinLevel zapcore.Level
	traceLevel     zapcore.Level
	ws             zapcore.WriteSyncer
}

var _ core.ILogger = (*logCore)(nil)

func newLogCore(conf *core.LogConfig, log *zap.Logger, ws zapcore.WriteSyncer) *logCore {
	l := &logCore{
		conf:           conf,
		log:            log,
		fields:         nil,
		level:          parserLogLevel(Level(conf.Level)),
		callerMinLevel: parserLogLevel(Level(conf.ShowFileAndLinenumMinLevel)),
		traceLevel:     parserLogLevel(Level(conf.TraceLevel)),
		ws:             ws,
	}
	return l
}

type logBody struct {
	ctx                    context.Context
	msg                    string
	fields                 []zap.Field
	customCaller           *customCaller
	withoutAttachLog2Trace bool
}

func (l *logCore) print(level Level, v []interface{}) {
	body := l.makeBody(v)
	zapLevel := parserLogLevel(level)
	ce := l.log.Check(zapLevel, body.msg)
	if ce != nil {
		if body.customCaller != nil {
			ce.Caller.Function = body.customCaller.Fn
			ce.Caller.File = body.customCaller.File
			ce.Caller.Line = body.customCaller.Line
			ce.Caller.Defined = true
		}
		ce.Caller.Defined = l.conf.ShowFileAndLinenumMinLevel != "" && zapLevel >= l.callerMinLevel
		if zapLevel >= l.level {
			ce.Write(body.fields...)
		}
	}

	// 将log附到trace上
	if l.conf.TraceLevel != "" && zapLevel >= l.traceLevel {
		l.attachLog2Trace(ce, body)
	}
}

func (l *logCore) makeBody(v []interface{}) logBody {
	args := make([]interface{}, 0, len(v))
	fields := append([]zap.Field{}, l.fields...)
	var cCaller *customCaller
	var ctx context.Context
	var withoutALT bool
	for _, value := range v {
		switch val := value.(type) {
		case zap.Field:
			if val.Type == zapcore.ReflectType && val.Key == "caller" {
				caller, ok := val.Interface.(*customCaller)
				if ok {
					cCaller = caller
					continue
				}
			}
			fields = append(fields, val)
		case *zap.Field:
			if val.Type == zapcore.ReflectType && val.Key == "caller" {
				caller, ok := val.Interface.(*customCaller)
				if ok {
					cCaller = caller
					continue
				}
			}
			fields = append(fields, *val)
		case context.Context:
			traceID, spanID := utils.Trace.GetOTELTraceID(val)
			if traceID != "" {
				fields = append(fields, zap.String(logTraceIdKey, traceID))
			}
			if spanID != "" {
				fields = append(fields, zap.String(logTraceSpanIdKey, spanID))
			}
			ctx = val
		case string, bool,
			int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64:
			args = append(args, value)
		case withoutAttachLog2Trace:
			withoutALT = true
			continue
		default:
			fields = append(fields, zap.Any("logData", value))
		}
	}
	return logBody{
		ctx:                    ctx,
		msg:                    fmt.Sprint(args...),
		fields:                 fields,
		customCaller:           cCaller,
		withoutAttachLog2Trace: withoutALT,
	}
}

func (l *logCore) Log(level string, v ...interface{}) {
	l.print(Level(level), v)
}
func (l *logCore) Debug(v ...interface{}) {
	l.print(DebugLevel, v)
}
func (l *logCore) Info(v ...interface{}) {
	l.print(InfoLevel, v)
}
func (l *logCore) Warn(v ...interface{}) {
	l.print(WarnLevel, v)
}
func (l *logCore) Error(v ...interface{}) {
	l.print(ErrorLevel, v)
}
func (l *logCore) DPanic(v ...interface{}) {
	l.print(DPanicLevel, v)
}
func (l *logCore) Panic(v ...interface{}) {
	l.print(PanicLevel, v)
}
func (l *logCore) Fatal(v ...interface{}) {
	l.print(FatalLevel, v)
}

// 添加fields
func (l *logCore) AddFields(fields ...zap.Field) {
	l.fields = append(append([]zap.Field{}, l.fields...), fields...)
}

/*
为log移除一些field, 返回移除的个数

	count	移除key的个数, <1表示所有
*/
func (l *logCore) RemoveFields(count int, fieldKeys ...string) int {
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
func (l *logCore) NewSessionLogger(fields ...zap.Field) core.ILogger {
	return &logCore{
		conf:           l.conf,
		log:            l.log,
		fields:         append(append(append([]zap.Field{}, l.fields...), zap.String(logIdKey, l.nextLoggerId())), fields...),
		level:          l.level,
		callerMinLevel: l.callerMinLevel,
		traceLevel:     l.traceLevel,
		ws:             l.ws,
	}
}

// 创建一个带链路id的log
func (l *logCore) NewTraceLogger(ctx context.Context, fields ...zap.Field) core.ILogger {
	traceID, _ := utils.Trace.GetOTELTraceID(ctx)
	return &logCore{
		conf:           l.conf,
		log:            l.log,
		fields:         append(append(append([]zap.Field{}, l.fields...), zap.String(logTraceIdKey, traceID)), fields...),
		level:          l.level,
		callerMinLevel: l.callerMinLevel,
		traceLevel:     l.traceLevel,
		ws:             l.ws,
	}
}

func (l *logCore) attachLog2Trace(ce *zapcore.CheckedEntry, body logBody) {
	if body.ctx == nil || ce == nil || body.withoutAttachLog2Trace {
		return
	}

	attr := make([]utils.OtelSpanKV, 0, 4+len(body.fields))
	attr = append(attr,
		utils.OtelSpanKey("level").String(ce.Level.String()),
		utils.OtelSpanKey("message").String(ce.Message),
	)
	for _, f := range body.fields {
		attr = append(attr, utils.OtelSpanKey(f.Key).String(convertFieldValue(f)))
	}
	if ce.Caller.Defined {
		attr = append(attr,
			utils.OtelSpanKey("line").String(ce.Caller.File+":"+strconv.Itoa(ce.Caller.Line)),
			utils.OtelSpanKey("func").String(ce.Caller.Function),
		)
	}
	if ce.Stack != "" {
		attr = append(attr, utils.OtelSpanKey("stack").String(ce.Stack))
	}
	utils.Trace.CtxEvent(body.ctx, "Log", attr...)
}

type customCaller struct {
	Fn, File string
	Line     int
}

func WithCaller(fn, file string, line int) zap.Field {
	return zap.Reflect("caller", &customCaller{
		Fn:   fn,
		File: file,
		Line: line,
	})
}
