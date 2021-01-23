/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2020/7/2
   Description :
-------------------------------------------------
*/

package core

// 记录器
type ILogger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	DPanic(v ...interface{})
	Panic(v ...interface{})
	Fatal(v ...interface{})
	// 创建一个镜像log
	NewMirrorLogger(tag ...string) ILogger
}
