/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/20
   Description :
-------------------------------------------------
*/

package pkg

import (
	"context"

	"github.com/zly-app/zapp/consts"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
)

var Context = new(contextUtil)

type contextUtil struct{}

// 基于传入的标准context生成一个新的标准context并保存log
func (c *contextUtil) SaveLoggerToContext(ctx context.Context, log core.ILogger) context.Context {
	return context.WithValue(ctx, consts.LoggerSaveFieldKey, log)
}

// 从标准context中获取log
func (c *contextUtil) GetLoggerFromContext(ctx context.Context) (core.ILogger, bool) {
	value := ctx.Value(consts.LoggerSaveFieldKey)
	log, ok := value.(core.ILogger)
	return log, ok
}

// 从标准context中获取log, 如果失败会panic
func (c *contextUtil) MustGetLoggerFromContext(ctx context.Context) core.ILogger {
	log, ok := c.GetLoggerFromContext(ctx)
	if !ok {
		logger.Log.Panic("can't load app_context from context")
	}
	return log
}
