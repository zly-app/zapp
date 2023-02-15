package utils

import (
	"context"
	"strings"

	"github.com/opentracing/opentracing-go"
)

var Trace = &traceCli{}

type traceCli struct{}

// 将span存入ctx
func (*traceCli) SaveSpan(ctx context.Context, span opentracing.Span) context.Context {
	return opentracing.ContextWithSpan(ctx, span)
}

// 从ctx中获取span
func (*traceCli) GetSpan(ctx context.Context) opentracing.Span {
	return opentracing.SpanFromContext(ctx)
}

// 开始一个span
func (*traceCli) StartSpan(operationName string) opentracing.Span {
	return opentracing.StartSpan(operationName)
}

// 获取或生成子span
func (*traceCli) GetChildSpan(ctx context.Context, operationName string) opentracing.Span {
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan != nil {
		return opentracing.StartSpan(operationName, opentracing.ChildOf(parentSpan.Context()))
	}
	return opentracing.StartSpan(operationName)
}

// 获取或生成跟随span
func (*traceCli) GetFollowSpan(ctx context.Context, operationName string) opentracing.Span {
	parentSpan := opentracing.SpanFromContext(ctx)
	if parentSpan != nil {
		return opentracing.StartSpan(operationName, opentracing.FollowsFrom(parentSpan.Context()))
	}
	return opentracing.StartSpan(operationName)
}

// 获取TraceID
func (c *traceCli) GetTraceID(span opentracing.Span) string {
	if span == nil {
		return ""
	}

	// 支持 otel 中获取 traceID
	{
		ctx := opentracing.ContextWithSpan(context.Background(), span)
		sc := Otel.GetSpan(ctx).SpanContext()
		if sc.IsValid() {
			return sc.TraceID().String()
		}
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

func (c *traceCli) GetTraceIDWithContext(ctx context.Context) (string, string) {
	// 支持 otel 中获取 traceID
	{
		sc := Otel.GetSpan(ctx).SpanContext()
		if sc.IsValid() {
			return sc.TraceID().String(), sc.SpanID().String()
		}
	}

	span := c.GetSpan(ctx)
	return c.GetTraceID(span), ""
}

func (*traceCli) getJaegerTraceID(carrier opentracing.TextMapCarrier) string {
	const TraceID = "uber-trace-id"
	values := strings.SplitN(carrier[TraceID], ":", 2)
	if len(values) >= 1 {
		return values[0]
	}
	return ""
}

func (*traceCli) getZipKinTraceID(carrier opentracing.TextMapCarrier) string {
	const TraceID = "x-b3-traceid"
	return carrier[TraceID]
}
