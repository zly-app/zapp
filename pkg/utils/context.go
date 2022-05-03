package utils

import (
	"context"

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
