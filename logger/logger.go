/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/20
   Description :
-------------------------------------------------
*/

package logger

import (
	"go.uber.org/zap"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
	"github.com/zly-app/zapp/pkg/zlog"
)

var Log core.ILogger = zlog.DefaultLogger

func NewLogger(appName string, c core.IConfig, opts ...zap.Option) core.ILogger {
	conf := c.Config().Frame.Log
	if utils.Reflect.IsZero(conf) {
		conf = zlog.DefaultConfig
		conf.Name = appName
	}
	if conf.Name == "" {
		conf.Name = appName
	}
	c.Config().Frame.Log = conf

	log := zlog.New(conf, opts...)
	Log = log
	return log
}

func Debug(v ...interface{}) {
	l := zlog.GetLogCore(Log)
	l.Debug(v...)
}
func Info(v ...interface{}) {
	l := zlog.GetLogCore(Log)
	l.Info(v...)
}
func Warn(v ...interface{}) {
	l := zlog.GetLogCore(Log)
	l.Warn(v...)
}
func Error(v ...interface{}) {
	l := zlog.GetLogCore(Log)
	l.Error(v...)
}
func Panic(v ...interface{}) {
	l := zlog.GetLogCore(Log)
	l.Panic(v...)
}
func Fatal(v ...interface{}) {
	l := zlog.GetLogCore(Log)
	l.Fatal(v...)
}
