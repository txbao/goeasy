package limiter

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

	"github.com/txbao/goeasy/config"
)

// Limiter HTTP 限流。
type Limiter struct {
	lim *rate.Limiter
}

func New(cfg config.LimiterCfg) *Limiter {
	if !cfg.Enabled {
		return &Limiter{lim: nil}
	}
	return &Limiter{lim: rate.NewLimiter(rate.Limit(cfg.QPS), cfg.Burst)}
}

func (l *Limiter) Allow() bool {
	if l == nil || l.lim == nil {
		return true
	}
	return l.lim.Allow()
}

// GinMiddleware 全局限流中间件。
func (l *Limiter) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if l != nil && l.lim != nil && !l.lim.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": http.StatusTooManyRequests,
				"msg":  "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}
