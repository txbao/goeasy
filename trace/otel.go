package trace

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/txbao/goeasy/config"
)

var (
	initOnce      sync.Once
	initErr       error
	tracerProvider *sdktrace.TracerProvider
	otelTracer    oteltrace.Tracer
	otelEnabled   bool
	serviceName   string
)

// Init 初始化追踪；exporter=otlp 时导出到 OTLP Collector（Tempo/Jaeger）。
func Init(cfg config.TraceCfg, appName string) error {
	initOnce.Do(func() {
		serviceName = appName
		if cfg.Service != "" {
			serviceName = cfg.Service
		}
		if serviceName == "" {
			serviceName = "goeasy"
		}
		if !cfg.Enabled || strings.ToLower(cfg.Exporter) != "otlp" || cfg.Endpoint == "" {
			otelEnabled = false
			return
		}
		ctx := context.Background()
		client, err := newOTLPClient(ctx, cfg)
		if err != nil {
			initErr = err
			return
		}
		exp, err := otlptrace.New(ctx, client)
		if err != nil {
			initErr = fmt.Errorf("otlp exporter: %w", err)
			return
		}
		res, err := resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceName(serviceName),
			),
		)
		if err != nil {
			initErr = fmt.Errorf("otlp resource: %w", err)
			return
		}
		sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))
		tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exp),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sampler),
		)
		otel.SetTracerProvider(tracerProvider)
		otelTracer = tracerProvider.Tracer("github.com/txbao/goeasy")
		otelEnabled = true
	})
	return initErr
}

func newOTLPClient(_ context.Context, cfg config.TraceCfg) (otlptrace.Client, error) {
	endpoint := strings.TrimPrefix(cfg.Endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	switch strings.ToLower(cfg.Protocol) {
	case "http":
		opts := []otlptracehttp.Option{otlptracehttp.WithEndpoint(endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		return otlptracehttp.NewClient(opts...), nil
	default:
		opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(endpoint)}
		if cfg.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		return otlptracegrpc.NewClient(opts...), nil
	}
}

// OTelEnabled 是否启用 OTLP 导出。
func OTelEnabled() bool {
	return otelEnabled
}

// Shutdown 优雅关闭 TracerProvider，刷新未发送 span。
func Shutdown(ctx context.Context) error {
	if tracerProvider == nil {
		return nil
	}
	return tracerProvider.Shutdown(ctx)
}

// startOTel 创建 OTel span 并同步本地 trace_id 供日志使用。
func startOTel(ctx context.Context, name string) (context.Context, func()) {
	if otelTracer == nil {
		return startLocal(ctx, name)
	}
	ctx2, span := otelTracer.Start(ctx, name)
	local := Span{
		TraceID: span.SpanContext().TraceID().String(),
		SpanID:  span.SpanContext().SpanID().String(),
		Name:    name,
	}
	ctx2 = context.WithValue(ctx2, ctxKey{}, local)
	span.SetAttributes(attribute.String("goeasy.service", serviceName))
	return ctx2, func() { span.End() }
}
