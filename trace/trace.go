package trace

import (
	"context"

	"github.com/google/uuid"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type ctxKey struct{}

// Span 链路追踪元数据（兼容本地与 OTel）。
type Span struct {
	TraceID string
	SpanID  string
	Name    string
}

// Start 开始 span；OTLP 启用时导出到 Collector，否则仅本地 trace_id。
func Start(ctx context.Context, name string) (context.Context, func()) {
	if otelEnabled {
		return startOTel(ctx, name)
	}
	return startLocal(ctx, name)
}

func startLocal(ctx context.Context, name string) (context.Context, func()) {
	span := Span{
		TraceID: traceIDFromCtx(ctx),
		SpanID:  uuid.NewString(),
		Name:    name,
	}
	if span.TraceID == "" {
		span.TraceID = uuid.NewString()
	}
	ctx = context.WithValue(ctx, ctxKey{}, span)
	return ctx, func() {}
}

func traceIDFromCtx(ctx context.Context) string {
	if sc := oteltrace.SpanFromContext(ctx).SpanContext(); sc.HasTraceID() {
		return sc.TraceID().String()
	}
	if v, ok := ctx.Value(ctxKey{}).(Span); ok {
		return v.TraceID
	}
	return ""
}

// TraceID 从 context 读取 trace_id。
func TraceID(ctx context.Context) string {
	return traceIDFromCtx(ctx)
}
