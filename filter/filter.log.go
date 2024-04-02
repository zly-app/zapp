package filter

import (
	"context"
	"time"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/logger"
	"github.com/zly-app/zapp/pkg/utils"
	"github.com/zly-app/zapp/pkg/zlog"
)

const defLogLevel = "debug"

func init() {
	RegisterFilterCreator("base.log", newLogFilter, newLogFilter)
}

var defLogFilter core.Filter = &logFilter{
	Client:  make(map[string]map[string]string),
	Service: make(map[string]string),
}

func newLogFilter() core.Filter {
	return defLogFilter
}

type logFilter struct {
	Client  map[string]map[string]string
	Service map[string]string
}

func (t *logFilter) getMethodName(meta CallMeta) string {
	return meta.CalleeMethod() + "/" + meta.CalleeMethod()
}

func (t *logFilter) marshal(a any) string {
	s, _ := sonic.MarshalString(a)
	return s
}

func (t *logFilter) Init(app core.IApp) error {
	err := config.Conf.ParseFilterConfig("base.log", t, true)
	if err != nil {
		return err
	}
	return nil
}

func (t *logFilter) getClientLevel(clientType, clientName string) string {
	ct, ok := t.Client[clientType]
	if !ok { // 没有找到 clientType 则用全局默认
		ct, ok = t.Client[defName]
		if ok {
			return ct[defName]
		}
		return defLogLevel
	}

	l, ok := ct[clientName]
	if ok {
		return l
	}
	return ct[defName]
}
func (t *logFilter) getServiceLevel(serviceName string) string {
	l, ok := t.Service[serviceName]
	if ok {
		return l
	}
	return t.Service[defName]
}
func (t *logFilter) getLevel(ctx context.Context) string {
	meta := GetCallMeta(ctx)
	l := defLogLevel
	if meta.IsClientMeta() {
		l = t.getClientLevel(meta.ClientType(), meta.ClientName())
	} else if meta.IsServiceMeta() {
		l = t.getServiceLevel(meta.ServiceName())
	}
	return l
}

func (t *logFilter) start(ctx context.Context, req interface{}) (context.Context, CallMeta, []interface{}) {
	meta := GetCallMeta(ctx)

	eventName := " Send"
	if meta.IsServiceMeta() {
		eventName = " Recv"
	}
	fn, file, line := meta.FuncFileLine()
	customCaller := zlog.WithCaller(fn, file, line)

	instance := zap.String("instance", config.Conf.Config().Frame.Instance)
	callerService := zap.String("callerService", meta.CallerService())
	callerMethod := zap.String("callerMethod", meta.CallerMethod())
	calleeService := zap.String("calleeService", meta.CalleeService())
	calleeMethod := zap.String("calleeMethod", meta.CalleeMethod())

	logFields := []interface{}{
		customCaller,
		instance,
		callerService,
		callerMethod,
		calleeService,
		calleeMethod,
	}

	level := t.getLevel(ctx)
	logger.Log.Log(level,
		customCaller,
		instance,
		callerService,
		callerMethod,
		calleeService,
		calleeMethod,
		ctx, t.getMethodName(meta)+eventName, zap.String("req", t.marshal(req)),
	)
	return ctx, meta, logFields
}

func (t *logFilter) end(ctx context.Context, meta CallMeta, rsp interface{}, err error, logFields []interface{}) error {
	code, codeType, replaceErr := DefaultGetErrCodeFunc(ctx, rsp, err)
	err = replaceErr

	eventName := " Recv"
	if meta.IsServiceMeta() {
		eventName = " Send"
	}

	duration := meta.EndTime() - meta.StartTime()
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
		level := t.getLevel(ctx)
		logger.Log.Log(level, logFields...)
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
