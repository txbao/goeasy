package sqllog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/metrics"
)

var (
	cfgEnabled bool
	cfgSlowMs  int
)

// Init 从配置初始化 SQL 日志（app 启动时调用）。
func Init(cfg config.SQLCfg) {
	cfgEnabled = cfg.Enabled
	cfgSlowMs = cfg.SlowMs
	if cfgSlowMs <= 0 {
		cfgSlowMs = 200
	}
}

// Enabled 是否输出 SQL。
func Enabled() bool {
	if cfgEnabled {
		return true
	}
	if v := os.Getenv("GOEASY_LOG_SQL"); v == "1" || strings.EqualFold(v, "true") {
		return true
	}
	if v := os.Getenv("GOEASY_LOG_SQL"); v == "0" || strings.EqualFold(v, "false") {
		return false
	}
	env := strings.ToLower(os.Getenv("GOEASY_ENV"))
	return env == "" || env == "dev" || env == "development" || env == "local"
}

// MaybeLog 在启用时打印 SQL 与参数，并标记慢查询。
func MaybeLog(ctx context.Context, op, query string, args ...any) {
	if !Enabled() {
		return
	}
	start := time.Now()
	slog.InfoContext(ctx, "sql "+op,
		"query", query,
		"args", fmt.Sprint(args...),
		"interpolated", interpolate(query, args),
	)
	elapsed := time.Since(start)
	if cfgSlowMs > 0 && elapsed.Milliseconds() >= int64(cfgSlowMs) {
		slog.WarnContext(ctx, "sql slow query",
			"op", op,
			"query", query,
			"duration_ms", elapsed.Milliseconds(),
			"slow_threshold_ms", cfgSlowMs,
		)
	}
}

// LogDuration 记录已执行 SQL 的耗时（供包装层在 Exec 后调用）。
func LogDuration(ctx context.Context, op, query string, elapsed time.Duration, args ...any) {
	if !Enabled() {
		if cfgSlowMs > 0 && elapsed.Milliseconds() >= int64(cfgSlowMs) {
			slog.WarnContext(ctx, "sql slow query",
				"op", op,
				"query", query,
				"duration_ms", elapsed.Milliseconds(),
				"slow_threshold_ms", cfgSlowMs,
				"args", fmt.Sprint(args...),
			)
		}
		return
	}
	level := slog.LevelInfo
	attrs := []any{
		"op", op,
		"query", query,
		"args", fmt.Sprint(args...),
		"duration_ms", elapsed.Milliseconds(),
	}
	if cfgSlowMs > 0 && elapsed.Milliseconds() >= int64(cfgSlowMs) {
		level = slog.LevelWarn
		attrs = append(attrs, "slow", true, "slow_threshold_ms", cfgSlowMs)
	}
	slog.Log(ctx, level, "sql", attrs...)
	slow := cfgSlowMs > 0 && elapsed.Milliseconds() >= int64(cfgSlowMs)
	metrics.ObserveSQL(op, slow)
}

func interpolate(query string, args []any) string {
	if len(args) == 0 {
		return query
	}
	var b strings.Builder
	b.WriteString(query)
	b.WriteString(" /* args: ")
	b.WriteString(fmt.Sprint(args...))
	b.WriteString(" */")
	return b.String()
}
