/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/4/17
   Description :
-------------------------------------------------
*/

package zlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type interceptorFunc func(ent *zapcore.Entry, fields []zapcore.Field) (cancel bool)

type interceptor struct {
	zapcore.Core
	funcs []interceptorFunc
}

// 拦截器
func WithHook(funcs ...interceptorFunc) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &interceptor{
			Core:  core,
			funcs: append(([]interceptorFunc)(nil), funcs...),
		}
	})
}
func (c *interceptor) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}
func (h *interceptor) With(fields []zapcore.Field) zapcore.Core {
	return &interceptor{
		Core:  h.Core.With(fields),
		funcs: h.funcs,
	}
}
func (c *interceptor) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	for i := range c.funcs {
		if c.funcs[i](&ent, fields) {
			return nil
		}
	}
	return c.Core.Write(ent, fields)
}
