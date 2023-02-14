package utils

import (
	"context"

	"github.com/spf13/cast"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var Otel = &otelCli{}

type otelCli struct{}

// key, 示例 OtelSpanKey("foo").String("bar")
type OtelSpanKey = attribute.Key
type OtelSpanKV = attribute.KeyValue

func (*otelCli) SaveToContext(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}

// 从ctx中获取span
func (*otelCli) GetSpan(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// 开始一个span
func (*otelCli) StartSpan(ctx context.Context, spanName string, attributes ...OtelSpanKV) (
	context.Context, trace.Span) {
	return otel.Tracer("").Start(ctx, spanName, trace.WithAttributes(attributes...))
}

// 设置span属性. 属性是作为元数据应用于跨度的键和值，可用于聚合、过滤和分组跟踪
func (*otelCli) SetSpanAttributes(span trace.Span, attributes ...OtelSpanKV) {
	span.SetAttributes(attributes...)
}

// 添加事件, 事件的属性不会用于聚合、过滤和分组跟踪
func (*otelCli) AddSpanEvent(span trace.Span, eventName string, attributes ...OtelSpanKV) {
	span.AddEvent(eventName, trace.WithAttributes(attributes...))
}

// 将span标记为错误
func (*otelCli) MarkSpanAnError(span trace.Span, isErr bool) {
	span.SetStatus(codes.Error, cast.ToString(isErr))
}

// 结束一个span
func (*otelCli) EndSpan(span trace.Span) {
	span.End()
}

// 获取 traceID
func (*otelCli) GetOTELTraceID(ctx context.Context) (traceID string, spanID string) {
	sc := trace.SpanContextFromContext(ctx)
	return sc.TraceID().String(), sc.SpanID().String()
}
