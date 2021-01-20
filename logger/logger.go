/*
-------------------------------------------------
   Author :       zlyuancn
   date：         2021/1/20
   Description :
-------------------------------------------------
*/

package logger

import (
	"sync/atomic"

	"github.com/zlyuancn/zlog"
	"github.com/zlyuancn/zutils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zly-app/zapp/core"
)

var Log core.ILogger = &logger{Loger: zlog.DefaultLogger}

type logger struct {
	zlog.Loger
}

const logIdKey = "logId"

var loggerId uint32

// 获取下个日志id
//
// 将数值转为32进制, 因为求余2的次幂可以用位运算所以采用 数字+22位英文字母
func (l *logger) nextLoggerId() string {
	id := atomic.AddUint32(&loggerId, 1)
	bs := []byte{48, 48, 48, 48, 48, 48, 48}

	i := len(bs) - 1
	for {
		bs[i] = byte(31&id) + 48 // 从字符0开始
		if bs[i] > 57 {          // 超过数字用字母表示
			bs[i] += 39
		}
		if id < 32 {
			return string(bs)
		}
		i--
		id >>= 5
	}
}

func withColoursMessageOfLoggerId(enable *bool) zap.Option {
	return zlog.WithHook(func(ent *zapcore.Entry, fields []zapcore.Field) (cancel bool) {
		if !*enable || ent.Message == "" {
			return
		}

		for _, field := range fields {
			if field.Key == logIdKey {
				ent.Message = makeColorMessageOfLoggerId(field.String, ent.Message)
				break
			}
		}
		return
	})
}

func makeColorMessageOfLoggerId(logId string, message string) string {
	var id uint32
	for _, c := range logId {
		id <<= 5
		if c >= 'a' {
			id += uint32(c) - 87
		} else {
			id += uint32(c) - 48
		}
	}

	color := zutils.ColorType(id&7) + zutils.Color.Default
	return zutils.Color.MakeColorText(color, message)
}

func (l *logger) NewMirror(tag ...string) core.ILogger {
	log, _ := zlog.WrapZapFieldsWithLoger(l.Loger, zap.String(logIdKey, l.nextLoggerId()), zap.Strings("logTag", tag))
	return &logger{Loger: log}
}

func NewLogger(appName string, c core.IConfig) core.ILogger {
	conf := c.Config().Frame
	if zutils.Reflect.IsZero(conf.Log) {
		conf.Log = zlog.DefaultConfig
		conf.Log.Name = appName
	}
	if conf.Log.Name == "" {
		conf.Log.Name = appName
	}
	c.Config().Frame.Log = conf.Log

	log := zlog.New(conf.Log, withColoursMessageOfLoggerId(&c.Config().Frame.Log.IsTerminal))
	return &logger{Loger: log}
}
