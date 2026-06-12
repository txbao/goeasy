package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/txbao/goeasy/config"
)

// Logger 统一日志（slog）。
type Logger struct {
	inner *slog.Logger
}

// NewFromSlog 从已有 slog.Logger 构造（测试或自定义 Handler 注入）。
func NewFromSlog(l *slog.Logger) *Logger {
	if l == nil {
		return New(nil)
	}
	return &Logger{inner: l}
}

func New(cfg *config.Config) *Logger {
	lc := config.LoggerCfg{}
	service := "goeasy"
	env := ""
	if cfg != nil {
		lc = cfg.Observability.Logger
		if cfg.AppName != "" {
			service = cfg.AppName
		}
		env = cfg.Env
	}
	level := parseLevel(lc.Level, env)
	var w io.Writer = os.Stdout
	if strings.EqualFold(lc.Output, "stderr") {
		w = os.Stderr
	}
	var h slog.Handler
	opts := &slog.HandlerOptions{Level: level, AddSource: level == slog.LevelDebug}
	if strings.EqualFold(lc.Format, "text") {
		h = slog.NewTextHandler(w, opts)
	} else {
		h = slog.NewJSONHandler(w, opts)
	}
	inner := slog.New(h).With("service", service)
	if env != "" {
		inner = inner.With("env", env)
	}
	return &Logger{inner: inner}
}

func parseLevel(raw, env string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info":
		return slog.LevelInfo
	default:
		if env == "dev" {
			return slog.LevelDebug
		}
		return slog.LevelInfo
	}
}

func (l *Logger) Slog() *slog.Logger {
	if l == nil || l.inner == nil {
		return slog.Default()
	}
	return l.inner
}

func (l *Logger) Infof(format string, args ...any) {
	l.Slog().Info(fmt.Sprintf(format, args...))
}

func (l *Logger) Errorf(format string, args ...any) {
	l.Slog().Error(fmt.Sprintf(format, args...))
}

func (l *Logger) WithTraceID(traceID string) *Logger {
	return &Logger{inner: l.Slog().With("trace_id", traceID)}
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{inner: l.Slog().With(args...)}
}
