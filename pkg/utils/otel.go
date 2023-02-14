package utils

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

var Otel = &otelCli{}

type otelCli struct{}

// 从ctx中获取span
func (*otelCli) GetSpan(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func (*otelCli) GetOTELTraceID(ctx context.Context) (traceID string, spanID string) {
	sc := trace.SpanContextFromContext(ctx)
	return sc.TraceID().String(), sc.SpanID().String()
}
