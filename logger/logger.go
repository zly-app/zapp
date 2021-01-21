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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
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

func withColoursMessageOfLoggerId() zap.Option {
	return zlog.WithHook(func(ent *zapcore.Entry, fields []zapcore.Field) (cancel bool) {
		if ent.Message == "" {
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

	color := ColorType(id&7) + defaultColor
	return makeColorText(color, message)
}

func (l *logger) NewMirrorLogger(tag ...string) core.ILogger {
	log, _ := zlog.WrapZapFieldsWithLoger(l.Loger, zap.String(logIdKey, l.nextLoggerId()), zap.Strings("logTag", tag))
	return &logger{Loger: log}
}

func NewLogger(appName string, c core.IConfig) core.ILogger {
	conf := c.Config().Frame.Log
	if utils.Reflect.IsZero(conf) {
		conf = zlog.DefaultConfig
		conf.Name = appName
	}
	if conf.Name == "" {
		conf.Name = appName
	}
	c.Config().Frame.Log = conf

	opts := []zap.Option{}
	if conf.IsTerminal {
		opts = append(opts, withColoursMessageOfLoggerId())
	}
	log := zlog.New(conf, opts...)
	return &logger{Loger: log}
}
