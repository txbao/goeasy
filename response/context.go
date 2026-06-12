package response

import (
	"github.com/gin-gonic/gin"

	"github.com/txbao/goeasy/contextx"
	"github.com/txbao/goeasy/trace"
)

type logFields struct {
	Method    string
	Path      string
	RequestID string
	TraceID   string
	UserID    string
}

func extractLogFields(c *gin.Context) logFields {
	if c == nil {
		return logFields{}
	}
	return logFields{
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		RequestID: c.GetString("request_id"),
		TraceID:   trace.TraceID(c.Request.Context()),
		UserID:    contextx.UserID(c.Request.Context()),
	}
}
