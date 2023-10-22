package filter

import (
	"context"
	"strings"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/zlog"
)

var _ core.Filter = (*LogFilter)(nil)

func init() {
	RegisterFilterCreator("log", func() core.Filter {
		return &LogFilter{}
	}, func() core.Filter {
		return &LogFilter{}
	})
}

type LogFilterConfig struct {
	Level string
}

type LogFilter struct {
	level string
}

func (t *LogFilter) getMethodName(meta *Meta) string {
	if meta.isClientMeta {
		return meta.ClientType + "/" + meta.ClientName + "/" + meta.MethodName
	}
	return meta.ServiceName + "/" + meta.MethodName
}

func (t *LogFilter) marshal(a any) string {
	s, _ := sonic.MarshalString(a)
	return s
}

func (t *LogFilter) Init() error {
	conf := &LogFilterConfig{}
	err := config.Conf.ParseFilterConfig("log", conf, true)
	if err != nil {
		return err
	}

	level := zlog.InfoLevel
	switch strings.ToLower(conf.Level) {
	case "debug":
		level = zlog.DebugLevel
	case "info":
		level = zlog.InfoLevel
	case "warn":
		level = zlog.WarnLevel
	case "error":
		level = zlog.ErrorLevel
	case "dpanic":
		level = zlog.DPanicLevel
	case "panic":
		level = zlog.PanicLevel
	case "fatal":
		level = zlog.FatalLevel
	}
	t.level = level
	return nil
}

func (t *LogFilter) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	meta := GetMetaFromCtx(ctx)
	fn, file, line := meta.FuncFileLine()
	customCaller := zlog.WithCaller(fn, file, line)
	if meta.isClientMeta {
		logger.Log.Log(t.level, ctx, t.getMethodName(meta)+" Send", zap.String("req", t.marshal(req)), customCaller)
	} else {
		logger.Log.Log(t.level, ctx, t.getMethodName(meta)+" Recv", zap.String("req", t.marshal(req)), customCaller)
	}

	err := next(ctx, req, rsp)

	if meta.isClientMeta {
		logger.Log.Log(t.level, ctx, t.getMethodName(meta)+" Recv", zap.String("rsp", t.marshal(rsp)), customCaller)
	} else {
		logger.Log.Log(t.level, ctx, t.getMethodName(meta)+" Send", zap.String("rsp", t.marshal(rsp)), customCaller)
	}
	return err
}

func (t *LogFilter) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (interface{}, error) {
	meta := GetMetaFromCtx(ctx)
	fn, file, line := meta.FuncFileLine()
	customCaller := zlog.WithCaller(fn, file, line)
	if meta.isClientMeta {
		logger.Log.Log(t.level, ctx, t.getMethodName(meta)+" Send", zap.String("req", t.marshal(req)), customCaller)
	} else {
		logger.Log.Log(t.level, ctx, t.getMethodName(meta)+" Recv", zap.String("req", t.marshal(req)), customCaller)
	}

	rsp, err := next(ctx, req)

	if meta.isClientMeta {
		logger.Log.Log(t.level, ctx, t.getMethodName(meta)+" Recv", zap.String("rsp", t.marshal(rsp)), customCaller)
	} else {
		logger.Log.Log(t.level, ctx, t.getMethodName(meta)+" Send", zap.String("rsp", t.marshal(rsp)), customCaller)
	}
	return rsp, err
}

func (t *LogFilter) Close() error { return nil }
