/*
-------------------------------------------------
   Author :       zlyuan
   date：         2019/8/30
   Description :
-------------------------------------------------
*/

package zlog

import (
	"go.uber.org/zap/zapcore"

	"github.com/zly-app/zapp/core"
)

type Level string

const (
	DebugLevel  = "debug"  // 开发用, 生产模式下不应该是debug
	InfoLevel   = "info"   // 默认级别, 用于告知程序运行情况
	WarnLevel   = "warn"   // 比信息更重要, 但不需要单独的人工检查
	ErrorLevel  = "error"  // 高优先级的, 如果应用程序运行正常, 就不应该生成任何错误级别的日志
	DPanicLevel = "dpanic" // 严重的错误, 在开发者模式下日志记录器在写完消息后程序会感到恐慌
	PanicLevel  = "panic"  // 记录一条消息, 然后记录一条消息, 然后程序会感到恐慌
	FatalLevel  = "fatal"  // 记录一条消息, 然后调用 os.Exit(1)
)

var levelMapping = map[Level]zapcore.Level{
	DebugLevel:  zapcore.DebugLevel,
	InfoLevel:   zapcore.InfoLevel,
	WarnLevel:   zapcore.WarnLevel,
	ErrorLevel:  zapcore.ErrorLevel,
	DPanicLevel: zapcore.DPanicLevel,
	PanicLevel:  zapcore.PanicLevel,
	FatalLevel:  zapcore.FatalLevel,
}

var levelMappingReverse = map[zapcore.Level]Level{
	zapcore.DebugLevel:  DebugLevel,
	zapcore.InfoLevel:   InfoLevel,
	zapcore.WarnLevel:   WarnLevel,
	zapcore.ErrorLevel:  ErrorLevel,
	zapcore.DPanicLevel: DPanicLevel,
	zapcore.PanicLevel:  PanicLevel,
	zapcore.FatalLevel:  FatalLevel,
}

var DefaultConfig = core.LogConfig{
	Level:                      "debug",
	Json:                       false,
	WriteToStream:              true,
	WriteToFile:                false,
	Name:                       "zlog",
	AppendPid:                  false,
	Path:                       "./log",
	FileMaxSize:                32,
	FileMaxBackupsNum:          3,
	FileMaxDurableTime:         7,
	TimeFormat:                 "2006-01-02 15:04:05",
	Color:                      true,
	CapitalLevel:               false,
	DevelopmentMode:            true,
	ShowFileAndLinenum:         true,
	ShowFileAndLinenumMinLevel: "debug", // 推荐所有等级都打印代码行, 相对于能快速定位问题来说, 这点性能损耗无关紧要
	MillisDuration:             true,
}
