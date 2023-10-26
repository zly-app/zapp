package filter

import (
	"context"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"

	"github.com/zly-app/zapp/config"
	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

var _ core.Filter = (*traceFilter)(nil)

func init() {
	RegisterFilterCreator("trace", NewTraceFilter, NewTraceFilter)
}

var defTraceFilter core.Filter = traceFilter{}

func NewTraceFilter() core.Filter {
	return defTraceFilter
}

type traceFilter struct {
}

func (t traceFilter) getSpanName(meta *CallMeta) string {
	return meta.CalleeService + "/" + meta.CalleeMethod
}
func (t traceFilter) marshal(a any) string {
	s, _ := sonic.MarshalString(a)
	return s
}

func (t traceFilter) Init() error { return nil }

func (t traceFilter) start(ctx context.Context, req interface{}) (context.Context, trace.Span, *CallMeta) {
	meta := GetCallMeta(ctx)
	fn, file, line := meta.FuncFileLine()

	kind := trace.SpanKindClient
	if !meta.isClientMeta {
		kind = trace.SpanKindServer
	}
	ctx, span := otel.Tracer("").Start(ctx, t.getSpanName(meta),
		trace.WithAttributes(
			utils.OtelSpanKey("line").String(file+":"+strconv.Itoa(line)),
			utils.OtelSpanKey("func").String(fn),
			utils.OtelSpanKey("instance").String(config.Conf.Config().Frame.Instance),
			utils.OtelSpanKey("calleeService").String(meta.CalleeService),
			utils.OtelSpanKey("calleeMethod").String(meta.CalleeMethod),
		),
		trace.WithSpanKind(kind),
	)

	eventName := "Send"
	if !meta.isClientMeta {
		eventName = "Recv"
	}
	utils.Otel.CtxEvent(ctx, eventName, utils.OtelSpanKey("req").String(t.marshal(req)))
	return ctx, span, meta
}

func (t traceFilter) end(ctx context.Context, span trace.Span, meta *CallMeta, rsp interface{}, err error) error {
	code, codeType, replaceErr := DefaultGetErrCodeFunc(ctx, rsp, err)
	err = replaceErr

	// 计时
	duration := meta.GetEndTime() - meta.GetStartTime()
	span.SetAttributes(
		utils.OtelSpanKey("duration").Int64(duration),
		utils.OtelSpanKey("durationText").String(time.Duration(duration).String()),
		utils.OtelSpanKey("code").Int(code),
		utils.OtelSpanKey("codeType").String(codeType),
	)

	eventName := "Recv"
	if !meta.isClientMeta {
		eventName = "Send"
	}

	if err != nil {
		utils.Otel.CtxErrEvent(ctx, eventName, err, utils.OtelSpanKey("rsp").String(t.marshal(rsp)))
	} else {
		utils.Otel.CtxEvent(ctx, eventName, utils.OtelSpanKey("rsp").String(t.marshal(rsp)))
	}
	return err
}

func (t traceFilter) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	ctx, span, meta := t.start(ctx, req)
	defer span.End()

	err := next(ctx, req, rsp)
	err = t.end(ctx, span, meta, rsp, err)
	return err
}

func (t traceFilter) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (interface{}, error) {
	ctx, span, meta := t.start(ctx, req)
	defer span.End()

	rsp, err := next(ctx, req)
	err = t.end(ctx, span, meta, rsp, err)
	return rsp, err
}

func (t traceFilter) Close() error { return nil }
