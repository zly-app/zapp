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
	core zapcore.Core
	conf *hookConfig
}

func (c *interceptor) Enabled(level zapcore.Level) bool {
	return c.core.Enabled(level)
}
func (c *interceptor) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}
func (h *interceptor) With(fields []zapcore.Field) zapcore.Core {
	conf := *h.conf
	if conf.core != nil {
		conf.core = h.conf.core.With(fields)
	}
	return &interceptor{
		core: h.core.With(fields),
		conf: &conf,
	}
}
func (c *interceptor) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	for _, fn := range c.conf.fns {
		if fn(&ent, fields) {
			return nil
		}
	}

	if c.conf.core != nil {
		err := c.conf.core.Write(ent, fields)
		if err != nil {
			return err
		}
	}
	return c.core.Write(ent, fields)
}
func (i *interceptor) Sync() error {
	if i.conf.core != nil {
		err := i.conf.core.Sync()
		if err != nil {
			return err
		}
	}
	return i.core.Sync()
}

func (i *interceptor) WrapCore(core zapcore.Core) zapcore.Core {
	if i.core != nil {
		return i
	}
	i.core = core

	for _, fn := range i.conf.startHookCallbacks {
		fn()
	}
	return i
}

// 核心附加
func WithCore(core zapcore.Core) zap.Option {
	conf := NewHookConfig()
	conf.SetCore(core)
	return WithHookByConfig(conf)
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
	core               zapcore.Core
	fns                []interceptorFunc
	startHookCallbacks []func()
}

// 添加附加核心
func (h *hookConfig) SetCore(core zapcore.Core) *hookConfig {
	h.core = core
	return h
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
