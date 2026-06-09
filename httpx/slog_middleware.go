package httpx

import (
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
		inner.Slog().Info("http_access",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"request_id", c.GetString("request_id"),
			"trace_id", trace.TraceID(c.Request.Context()),
			"client_ip", c.ClientIP(),
			"user_id", contextx.UserID(c.Request.Context()),
		)
	}
}
