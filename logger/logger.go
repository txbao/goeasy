package logger

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/txbao/goeasy/config"
)

// Logger 统一日志（slog JSON）。
type Logger struct {
	inner *slog.Logger
}

func New(cfg *config.Config) *Logger {
	level := slog.LevelInfo
	if cfg != nil && cfg.Env == "dev" {
		level = slog.LevelDebug
	}
	service := "goesy"
	if cfg != nil && cfg.AppName != "" {
		service = cfg.AppName
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	inner := slog.New(h).With("service", service)
	return &Logger{inner: inner}
}

func (l *Logger) Infof(format string, args ...any) {
	l.inner.Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...any) {
	l.inner.Error(fmt.Sprintf(format, args...))
}

func (l *Logger) WithTraceID(traceID string) *Logger {
	return &Logger{inner: l.inner.With("trace_id", traceID)}
}
