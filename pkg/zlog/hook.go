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
	conf *hookConfig
}

func (c *interceptor) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}
func (h *interceptor) With(fields []zapcore.Field) zapcore.Core {
	return &interceptor{
		Core: h.Core.With(fields),
		conf: h.conf,
	}
}
func (c *interceptor) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	for _, fn := range c.conf.fns {
		if fn(&ent, fields) {
			return nil
		}
	}
	return c.Core.Write(ent, fields)
}

func (i *interceptor) WrapCore(core zapcore.Core) zapcore.Core {
	if i.Core != nil {
		return i
	}
	i.Core = core

	for _, fn := range i.conf.startHookCallbacks {
		fn()
	}
	return i
}

// 拦截器
func WithHook(fns ...interceptorFunc) zap.Option {
	conf := NewHookConfig()
	conf.AddInterceptorFunc(fns...)
	return WithHookByConfig(conf)
}

// 拦截器
func WithHookByConfig(conf *hookConfig) zap.Option {
	i := &interceptor{
		conf: conf,
	}
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return i.WrapCore(core)
	})
}

type hookConfig struct {
	fns                []interceptorFunc
	startHookCallbacks []func()
}

// 添加开始hook回调
func (h *hookConfig) AddStartHookCallbacks(callbacks ...func()) *hookConfig {
	h.startHookCallbacks = append(h.startHookCallbacks, callbacks...)
	return h
}

// 添加拦截器函数
func (h *hookConfig) AddInterceptorFunc(fns ...interceptorFunc) *hookConfig {
	h.fns = append(h.fns, fns...)
	return h
}

func NewHookConfig() *hookConfig {
	return &hookConfig{}
}
