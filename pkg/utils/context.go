package utils

import (
	"context"
	"time"

	"github.com/zly-app/zapp/core"
)

var Ctx = &ctxCli{}

type ctxCli struct{}

var ctxLoggerKey = &ctxLoggerField{}

type ctxLoggerField struct{}

// 将log存入ctx
func (*ctxCli) SaveLogger(ctx context.Context, log core.ILogger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey, log)
}

// 获取log, 不存在时返回nil
func (*ctxCli) GetLogger(ctx context.Context) core.ILogger {
	value := ctx.Value(ctxLoggerKey)
	log, ok := value.(core.ILogger)
	if ok {
		return log
	}
	return nil
}

// clone一个不会超时的ctx, 且携带上下文
func (*ctxCli) CloneContext(ctx context.Context) context.Context {
	return detach(ctx)
}

type detachedContext struct{ parent context.Context }

func detach(ctx context.Context) context.Context            { return detachedContext{ctx} }
func (v detachedContext) Deadline() (time.Time, bool)       { return time.Time{}, false }
func (v detachedContext) Done() <-chan struct{}             { return nil }
func (v detachedContext) Err() error                        { return nil }
func (v detachedContext) Value(key interface{}) interface{} { return v.parent.Value(key) }
