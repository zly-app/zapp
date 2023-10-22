package filter

import (
	"context"
	"strconv"

	"github.com/bytedance/sonic"

	"github.com/zly-app/zapp/core"
	"github.com/zly-app/zapp/pkg/utils"
)

var _ core.Filter = (*TraceFilter)(nil)

func init() {
	RegisterFilterCreator("trace", func() core.Filter {
		return TraceFilter{}
	}, func() core.Filter {
		return TraceFilter{}
	})
}

type TraceFilter struct {
}

func (t TraceFilter) getSpanName(meta *Meta) string {
	if meta.isClientMeta {
		return meta.ClientType + "/" + meta.ClientName + "/" + meta.MethodName
	}
	return meta.ServiceName + "/" + meta.MethodName
}
func (t TraceFilter) marshal(a any) string {
	s, _ := sonic.MarshalString(a)
	return s
}

func (t TraceFilter) Init() error { return nil }

func (t TraceFilter) HandleInject(ctx context.Context, req, rsp interface{}, next core.FilterInjectFunc) error {
	meta := GetMetaFromCtx(ctx)
	fn, file, line := meta.FuncFileLine()
	ctx, span := utils.Otel.StartSpan(ctx, t.getSpanName(meta),
		utils.OtelSpanKey("line").String(file+":"+strconv.Itoa(line)),
		utils.OtelSpanKey("func").String(fn),
	)
	defer span.End()

	if meta.isClientMeta {
		utils.Otel.CtxEvent(ctx, "Send", utils.OtelSpanKey("req").String(t.marshal(req)))
	} else {
		utils.Otel.CtxEvent(ctx, "Recv", utils.OtelSpanKey("req").String(t.marshal(req)))
	}

	err := next(ctx, req, rsp)

	if meta.isClientMeta {
		utils.Otel.CtxEvent(ctx, "Recv", utils.OtelSpanKey("rsp").String(t.marshal(rsp)))
	} else {
		utils.Otel.CtxEvent(ctx, "Send", utils.OtelSpanKey("rsp").String(t.marshal(rsp)))
	}
	return err
}

func (t TraceFilter) Handle(ctx context.Context, req interface{}, next core.FilterFunc) (interface{}, error) {
	meta := GetMetaFromCtx(ctx)
	fn, file, line := meta.FuncFileLine()
	ctx, span := utils.Otel.StartSpan(ctx, t.getSpanName(meta),
		utils.OtelSpanKey("line").String(file+":"+strconv.Itoa(line)),
		utils.OtelSpanKey("func").String(fn),
	)
	defer span.End()

	if meta.isClientMeta {
		utils.Otel.CtxEvent(ctx, "Send", utils.OtelSpanKey("req").String(t.marshal(req)))
	} else {
		utils.Otel.CtxEvent(ctx, "Recv", utils.OtelSpanKey("req").String(t.marshal(req)))
	}

	rsp, err := next(ctx, req)

	if meta.isClientMeta {
		utils.Otel.CtxEvent(ctx, "Recv", utils.OtelSpanKey("rsp").String(t.marshal(rsp)))
	} else {
		utils.Otel.CtxEvent(ctx, "Send", utils.OtelSpanKey("rsp").String(t.marshal(rsp)))
	}
	return rsp, err
}

func (t TraceFilter) Close() error { return nil }
