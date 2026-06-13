package httpx

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/txbao/goeasy/contextx"
	"github.com/txbao/goeasy/trace"
)

const maxDeviceInfoLen = 500

// InjectOperatorContext 将 gin 中的 JWT / 请求元数据写入 context.Context。
// 挂载顺序建议：requestID → trace → RequireJWT → InjectOperatorContext。
func InjectOperatorContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if rid := c.GetString("request_id"); rid != "" {
			ctx = contextx.WithRequestID(ctx, rid)
		}
		if tid := trace.TraceID(ctx); tid != "" {
			ctx = contextx.WithTraceID(ctx, tid)
		}

		subject, _ := c.Get("jwt_subject")
		if s, ok := subject.(string); ok && s != "" {
			ctx = contextx.WithUserID(ctx, s)
			ctx = contextx.WithLoginID(ctx, s)
		}
		if cid, ok := c.Get("jwt_customer_id"); ok {
			if v, ok := cid.(int64); ok {
				ctx = contextx.WithCustomerID(ctx, v)
			}
		}
		if pla, ok := c.Get("jwt_platform_admin"); ok {
			if v, ok := pla.(bool); ok {
				ctx = contextx.WithPlatformAdmin(ctx, v)
			}
		}

		ctx = contextx.WithClientIP(ctx, c.ClientIP())
		ctx = contextx.WithDeviceInfo(ctx, truncateDeviceInfo(c.GetHeader("User-Agent")))

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func truncateDeviceInfo(ua string) string {
	ua = strings.TrimSpace(ua)
	if len(ua) <= maxDeviceInfoLen {
		return ua
	}
	return ua[:maxDeviceInfoLen]
}
