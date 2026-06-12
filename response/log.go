package response

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func shouldLogServerError(httpCode, bizCode int, enabled bool) bool {
	if !enabled {
		return false
	}
	return httpCode >= 500 || bizCode >= 500000
}

func logServerError(c *gin.Context, httpCode, bizCode int, err error, msg string) {
	log, env, enabled, _ := snapshot()
	if !shouldLogServerError(httpCode, bizCode, enabled) {
		return
	}
	fields := extractLogFields(c)
	attrs := []any{
		"method", fields.Method,
		"path", fields.Path,
		"request_id", fields.RequestID,
		"trace_id", fields.TraceID,
		"user_id", fields.UserID,
		"biz_code", bizCode,
		"http_status", httpCode,
		"env", env,
	}
	if err != nil {
		attrs = append(attrs, "err", err.Error())
	} else if msg != "" {
		attrs = append(attrs, "err", msg)
	}
	sl := slog.Default()
	if log != nil {
		sl = log.Slog()
	}
	sl.Error("http_server_error", attrs...)
}
