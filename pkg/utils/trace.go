package utils

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var Trace = &otelCli{}

type otelCli struct{}

// key, 示例 OtelSpanKey("foo").String("bar")
type OtelSpanKey = attribute.Key
type OtelSpanKV = attribute.KeyValue

func (*otelCli) GlobalTrace(name string) trace.Tracer {
	return otel.Tracer(name)
}

// 将span存入ctx中
func (*otelCli) SaveSpan(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}

// 从ctx中获取span
func (*otelCli) GetSpan(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func (*otelCli) SaveToHeaders(ctx context.Context, headers http.Header) {
	v := propagation.HeaderCarrier(headers)
	otel.GetTextMapPropagator().Inject(ctx, v)
}

func (c *otelCli) GetSpanWithHeaders(ctx context.Context, headers http.Header) (context.Context, trace.Span) {
	v := propagation.HeaderCarrier(headers)
	ctx = otel.GetTextMapPropagator().Extract(ctx, v)
	return ctx, c.GetSpan(ctx)
}

func (*otelCli) SaveToMap(ctx context.Context, mapping map[string]string) {
	v := propagation.MapCarrier(mapping)
	otel.GetTextMapPropagator().Inject(ctx, v)
}

func (c *otelCli) GetSpanWithMap(ctx context.Context, mapping map[string]string) (context.Context, trace.Span) {
	v := propagation.MapCarrier(mapping)
	ctx = otel.GetTextMapPropagator().Extract(ctx, v)
	return ctx, c.GetSpan(ctx)
}

func (*otelCli) SaveToTextMapCarrier(ctx context.Context, carrier propagation.TextMapCarrier) {
	otel.GetTextMapPropagator().Inject(ctx, carrier)
}

func (c *otelCli) GetSpanWithTextMapCarrier(ctx context.Context, carrier propagation.TextMapCarrier) (context.Context, trace.Span) {
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	return ctx, c.GetSpan(ctx)
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
func (*otelCli) MarkSpanAnError(span trace.Span, err error) {
	span.SetStatus(codes.Error, err.Error())
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

// 根据超时ctx获取一个OtelSpanKV描述
func (*otelCli) GetSpanKVWithDeadline(ctx context.Context) OtelSpanKV {
	deadline, deadlineOK := ctx.Deadline()
	if !deadlineOK {
		return OtelSpanKey("ctx.deadline").Bool(false)
	}
	d := deadline.Sub(time.Now()) // 剩余时间
	return OtelSpanKey("ctx.deadline").String(d.String())
}

// 创建一个OtelSpanKV描述
func (*otelCli) AttrKey(key string) OtelSpanKey {
	return OtelSpanKey(key)
}

// 开始一个span
func (c *otelCli) CtxStart(ctx context.Context, name string, attributes ...OtelSpanKV) context.Context {
	// 生成新的 span
	ctx, _ = c.StartSpan(ctx, name, attributes...)
	return ctx
}

func (c *otelCli) CtxEvent(ctx context.Context, name string, attributes ...OtelSpanKV) {
	span := c.GetSpan(ctx)
	attr := []OtelSpanKV{c.GetSpanKVWithDeadline(ctx)}
	attr = append(attr, attributes...)
	c.AddSpanEvent(span, name, attr...)
}

func (c *otelCli) CtxErrEvent(ctx context.Context, name string, err error, attributes ...OtelSpanKV) {
	span := c.GetSpan(ctx)
	attr := []OtelSpanKV{
		c.GetSpanKVWithDeadline(ctx),
		OtelSpanKey("err.detail").String(err.Error()),
	}
	if Recover.IsRecoverError(err) {
		c.SetSpanAttributes(span, OtelSpanKey("panic").Bool(true))
		panicErrs := Recover.GetRecoverErrors(err)
		attr = append(attr, OtelSpanKey("detail").StringSlice(panicErrs), OtelSpanKey("panic").Bool(true))
	}

	attr = append(attr, attributes...)
	c.AddSpanEvent(span, name, attr...)
	c.MarkSpanAnError(span, err)
}

func (c *otelCli) CtxEnd(ctx context.Context) {
	span := c.GetSpan(ctx)
	c.EndSpan(span)
}
