package httpx

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/txbao/goeasy/config"
)

// CORSMiddleware 按配置注入 CORS 响应头（P1）。
func CORSMiddleware(cfg config.CORSCfg) gin.HandlerFunc {
	if !cfg.Enabled {
		return func(c *gin.Context) { c.Next() }
	}
	origins := cfg.AllowOrigins
	methods := cfg.AllowMethods
	headers := cfg.AllowHeaders
	if len(methods) == 0 {
		methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(headers) == 0 {
		headers = []string{"Content-Type", "Authorization", "X-Request-ID"}
	}
	allowMethods := strings.Join(methods, ", ")
	allowHeaders := strings.Join(headers, ", ")
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowOrigin(origins, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			if cfg.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		} else if len(origins) == 1 && origins[0] == "*" {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Methods", allowMethods)
		c.Header("Access-Control-Allow-Headers", allowHeaders)
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func allowOrigin(allowed []string, origin string) bool {
	for _, o := range allowed {
		if o == "*" || strings.EqualFold(o, origin) {
			return true
		}
	}
	return false
}
