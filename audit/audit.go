package audit

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/txbao/goeasy/config"
)

// Record 审计记录（无业务实体）。
type Record struct {
	Operator  string         `json:"operator"`
	Action    string         `json:"action"`
	Resource  string         `json:"resource"`
	IP        string         `json:"ip"`
	OldValue  map[string]any `json:"old_value,omitempty"`
	NewValue  map[string]any `json:"new_value,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

// Logger 审计日志写入。
type Logger struct {
	inner *slog.Logger
}

func New(cfg config.AuditCfg) *Logger {
	if !cfg.Enabled {
		return &Logger{inner: nil}
	}
	return &Logger{inner: slog.New(slog.NewJSONHandler(os.Stdout, nil))}
}

func (l *Logger) Record(ctx context.Context, rec Record) {
	if l == nil || l.inner == nil {
		return
	}
	if rec.Timestamp.IsZero() {
		rec.Timestamp = time.Now()
	}
	l.inner.InfoContext(ctx, "audit",
		"operator", rec.Operator,
		"action", rec.Action,
		"resource", rec.Resource,
		"ip", rec.IP,
	)
}
