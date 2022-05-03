package utils

import (
	"context"
	"strings"

	"github.com/opentracing/opentracing-go"
)

var Trace = &traceCli{}

type traceCli struct{}

// 将span存入ctx
func (*ctxCli) SaveSpan(ctx context.Context, span opentracing.Span) context.Context {
	return opentracing.ContextWithSpan(ctx, span)
}

// 获取或生成子span
func (*ctxCli) GetChildSpan(ctx context.Context, operationName string) opentracing.Span {
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan != nil {
		return opentracing.StartSpan(operationName, opentracing.ChildOf(parentSpan.Context()))
	}
	return opentracing.StartSpan(operationName)
}

// 获取或生成跟随span
func (*ctxCli) GetFollowSpan(ctx context.Context, operationName string) opentracing.Span {
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan != nil {
		return opentracing.StartSpan(operationName, opentracing.FollowsFrom(parentSpan.Context()))
	}
	return opentracing.StartSpan(operationName)
}

// 获取TraceID
func (c *ctxCli) GetTraceID(span opentracing.Span) string {
	if span == nil {
		return ""
	}
	trace := span.Tracer()
	if trace == nil {
		return ""
	}

	carrier := opentracing.TextMapCarrier{}
	err := trace.Inject(span.Context(), opentracing.TextMap, carrier)
	if err != nil {
		return ""
	}

	traceID := c.getJaegerTraceID(carrier)
	if traceID == "" {
		traceID = c.getZipKinTraceID(carrier)
	}
	return traceID
}

func (*ctxCli) getJaegerTraceID(carrier opentracing.TextMapCarrier) string {
	const TraceID = "uber-trace-id"
	values := strings.SplitN(carrier[TraceID], ":", 2)
	if len(values) >= 1 {
		return values[0]
	}
	return ""
}

func (*ctxCli) getZipKinTraceID(carrier opentracing.TextMapCarrier) string {
	const TraceID = "x-b3-traceid"
	return carrier[TraceID]
}
