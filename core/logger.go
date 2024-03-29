/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/7/2
   Description :
-------------------------------------------------
*/

package core

import (
	"context"

	"go.uber.org/zap"
)

// 记录器
type ILogger interface {
	Log(level string, v ...interface{})
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	DPanic(v ...interface{})
	Panic(v ...interface{})
	Fatal(v ...interface{})
	// 创建一个会话log
	NewSessionLogger(fields ...zap.Field) ILogger
	// 创建一个带链路id的log
	NewTraceLogger(ctx context.Context, fields ...zap.Field) ILogger
}
