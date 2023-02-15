package zlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zly-app/zapp/core"
)

var _ core.ILogger = (*logWrap)(nil)

type logWrap struct {
	*logCore
}

func newLogWrap(l *logCore) *logWrap {
	return &logWrap{l}
}

func (l *logWrap) Debug(v ...interface{}) {
	l.logCore.Debug(v...)
}

func (l *logWrap) Info(v ...interface{}) {
	l.logCore.Info(v...)
}

func (l *logWrap) Warn(v ...interface{}) {
	l.logCore.Warn(v...)
}

func (l *logWrap) Error(v ...interface{}) {
	l.logCore.Error(v...)
}

func (l *logWrap) DPanic(v ...interface{}) {
	l.logCore.DPanic(v...)
}

func (l *logWrap) Panic(v ...interface{}) {
	l.logCore.Panic(v...)
}

func (l *logWrap) Fatal(v ...interface{}) {
	l.logCore.Fatal(v...)
}

// 获取日志输出合成器
func GetLogWriteSyncer(l interface{}) (zapcore.WriteSyncer, bool) {
	switch a := l.(type) {
	case *logCore:
		return a.ws, true
	case *logWrap:
		return a.ws, true
	}
	return nil, false
}

// 为log添加一些field
func AddFields(l interface{}, fields ...zap.Field) bool {
	switch a := l.(type) {
	case *logCore:
		a.AddFields(fields...)
		return true
	case *logWrap:
		a.AddFields(fields...)
		return true
	}
	return false
}

/*
为log移除一些field, 返回移除的个数

	count	移除key的个数, <1表示所有
*/
func RemoveFields(l interface{}, count int, fieldKeys ...string) (int, bool) {
	switch a := l.(type) {
	case *logCore:
		return a.RemoveFields(count, fieldKeys...), true
	case *logWrap:
		return a.RemoveFields(count, fieldKeys...), true
	}
	return 0, false
}

// 获取logCore, logCore 打印的日志堆栈会-1, 使用者应该为其打印方法再包一层调用, 否则显示的文件行存在异常
func GetLogCore(l core.ILogger) core.ILogger {
	switch a := l.(type) {
	case *logCore:
		return l
	case *logWrap:
		return a.logCore
	}
	return l
}
