package httpx

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/txbao/goeasy/contextx"
	"github.com/txbao/goeasy/logger"
	"github.com/txbao/goeasy/trace"
)

// SlogAccessLog 结构化访问日志（Loki 友好）。
func SlogAccessLog(log *logger.Logger) gin.HandlerFunc {
	inner := log
	if inner == nil {
		inner = logger.New(nil)
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		status := c.Writer.Status()
		ctx := c.Request.Context()
		attrs := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latency_ms", time.Since(start).Milliseconds(),
			"request_id", contextx.RequestID(ctx),
			"trace_id", trace.TraceID(ctx),
			"client_ip", contextx.ClientIP(ctx),
			"user_id", contextx.UserID(ctx),
			"customer_id", contextx.CustomerID(ctx),
		}
		if status >= http.StatusInternalServerError {
			inner.Slog().Warn("http_access", attrs...)
		} else {
			inner.Slog().Info("http_access", attrs...)
		}
	}
}
