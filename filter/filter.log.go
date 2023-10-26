package filter

import (
	"context"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
	"github.com/zly-app/zapp/pkg/zlog"
)

var _ core.Filter = (*logFilter)(nil)

func init() {
	RegisterFilterCreator("log", NewLogFilter, NewLogFilter)
}

var defLogFilter core.Filter = &logFilter{}

func NewLogFilter() core.Filter {
	return defLogFilter
}

type logFilterConfig struct {
	Level string
}

type logFilter struct {
	level string
}

func (t *logFilter) getMethodName(meta *CallMeta) string {
	return meta.CalleeService + "/" + meta.CalleeMethod
}

func (t *logFilter) marshal(a any) string {
	s, _ := sonic.MarshalString(a)
	return s
}

func (t *logFilter) Init() error {
	conf := &logFilterConfig{}
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

func (t *logFilter) start(ctx context.Context, req interface{}) (context.Context, *CallMeta, []interface{}) {
	meta := GetCallMeta(ctx)

	eventName := " Send"
	if !meta.isClientMeta {
		eventName = " Recv"
	}
	fn, file, line := meta.FuncFileLine()
	customCaller := zlog.WithCaller(fn, file, line)

	instance := zap.String("instance", config.Conf.Config().Frame.Instance)
	calleeService := zap.String("calleeService", meta.CalleeService)
	calleeMethod := zap.String("calleeMethod", meta.CalleeMethod)

	logFields := []interface{}{
		customCaller,
		instance,
		calleeService,
		calleeMethod,
	}

	logger.Log.Log(t.level,
		customCaller,
		instance,
		calleeService,
		calleeMethod,
		ctx, t.getMethodName(meta)+eventName, zap.String("req", t.marshal(req)),
	)
	return ctx, meta, logFields
}

func (t *logFilter) end(ctx context.Context, meta *CallMeta, rsp interface{}, err error, logFields []interface{}) error {
	code, codeType, replaceErr := DefaultGetErrCodeFunc(ctx, rsp, err)
	err = replaceErr

	eventName := " Recv"
	if !meta.isClientMeta {
		eventName = " Send"
	}

	duration := meta.GetEndTime() - meta.GetStartTime()
	logFields = append(logFields,
		ctx,
		t.getMethodName(meta)+eventName,
		zap.String("rsp", t.marshal(rsp)),
		zap.Int64("duration", duration),
		zap.String("durationText", time.Duration(duration).String()),
		zap.Int("code", code),
		zap.String("codeType", codeType),
	)
	if err != nil {
		if meta.HasPanic() {
			detail := utils.Recover.GetRecoverErrors(err)
			logFields = append(logFields,
				zap.Error(err),
				zap.Bool("panic", true),
				zap.Strings("err.detail", detail),
			)
		} else {
			logFields = append(logFields, zap.Error(err))
		}
		logger.Log.Log(zlog.ErrorLevel, logFields...)
	} else {
		logger.Log.Log(t.level, logFields...)
	}
	return err
}

func (t *logFilter) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	ctx, meta, logFields := t.start(ctx, req)

	err := next(ctx, req, rsp)
	err = t.end(ctx, meta, rsp, err, logFields)
	return err
}

func (t *logFilter) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (interface{}, error) {
	ctx, meta, logFields := t.start(ctx, req)

	rsp, err := next(ctx, req)
	err = t.end(ctx, meta, rsp, err, logFields)
	return rsp, err
}

func (t *logFilter) Close() error { return nil }
