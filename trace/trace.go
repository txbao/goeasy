package trace

import (
	"context"

	"github.com/google/uuid"

	"github.com/txbao/goeasy/config"
)

type ctxKey struct{}

// Span 链路追踪占位（P3 可换 OpenTelemetry SDK）。
type Span struct {
	TraceID string
	SpanID  string
	Name    string
}

func Start(ctx context.Context, name string) (context.Context, func()) {
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
	if v, ok := ctx.Value(ctxKey{}).(Span); ok {
		return v.TraceID
	}
	return ""
}

// TraceID 从 context 读取 trace_id。
func TraceID(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKey{}).(Span); ok {
		return v.TraceID
	}
	return ""
}

func Init(cfg config.TraceCfg) {
	_ = cfg
}
