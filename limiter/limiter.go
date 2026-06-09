package limiter

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"

	"github.com/txbao/goeasy/config"
)

// Limiter HTTP 多维限流（本地 / Redis / 组合）。
type Limiter struct {
	enabled bool
	dims    []dimension
}

type dimension struct {
	name  string
	local *rate.Limiter
	redis *redisLimiter
}

// New 创建限流器；redis 可在 InitInfra 后通过 BindRedis 绑定。
func New(cfg config.LimiterCfg) *Limiter {
	if !cfg.Enabled {
		return &Limiter{}
	}
	return build(cfg, nil)
}

// BindRedis 在 Redis 连接就绪后绑定分布式限流。
func (l *Limiter) BindRedis(client *redis.Client) {
	if l == nil || client == nil {
		return
	}
	for i := range l.dims {
		if l.dims[i].redis != nil {
			l.dims[i].redis.client = client
		}
	}
}

func build(cfg config.LimiterCfg, rdb *redis.Client) *Limiter {
	mode := strings.ToLower(strings.TrimSpace(cfg.Mode))
	if mode == "" {
		mode = "local"
	}
	l := &Limiter{enabled: true}
	for _, dim := range cfg.Dimensions {
		name := strings.ToLower(strings.TrimSpace(dim))
		if name == "" {
			continue
		}
		d := dimension{name: name}
		if mode == "local" || mode == "both" {
			d.local = rate.NewLimiter(rate.Limit(cfg.QPS), cfg.Burst)
		}
		if (mode == "redis" || mode == "both") && rdb != nil {
			d.redis = &redisLimiter{
				client: rdb,
				qps:    cfg.QPS,
				burst:  cfg.Burst,
				prefix: "goeasy:rl:" + name,
			}
		} else if mode == "redis" || mode == "both" {
			d.redis = &redisLimiter{qps: cfg.QPS, burst: cfg.Burst, prefix: "goeasy:rl:" + name}
		}
		l.dims = append(l.dims, d)
	}
	if len(l.dims) == 0 {
		l.dims = append(l.dims, dimension{
			name:  "global",
			local: rate.NewLimiter(rate.Limit(cfg.QPS), cfg.Burst),
		})
	}
	return l
}

func (l *Limiter) Allow() bool {
	if l == nil || !l.enabled {
		return true
	}
	return l.allowKey("global", "")
}

func (l *Limiter) allowKey(dimName, key string) bool {
	for _, d := range l.dims {
		if d.name != dimName {
			continue
		}
		if d.local != nil && !d.local.Allow() {
			return false
		}
		if d.redis != nil && !d.redis.allow(key) {
			return false
		}
		return true
	}
	return true
}

func (l *Limiter) allowAll(c *gin.Context) bool {
	if l == nil || !l.enabled {
		return true
	}
	for _, d := range l.dims {
		key := dimensionKey(c, d.name)
		if d.local != nil && !d.local.Allow() {
			return false
		}
		if d.redis != nil && !d.redis.allow(key) {
			return false
		}
	}
	return true
}

func dimensionKey(c *gin.Context, dim string) string {
	switch dim {
	case "ip":
		return c.ClientIP()
	case "user":
		if sub, ok := c.Get("jwt_subject"); ok {
			if s, ok := sub.(string); ok && s != "" {
				return s
			}
		}
		return c.ClientIP()
	default:
		return "global"
	}
}

// GinMiddleware 多维限流中间件。
func (l *Limiter) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !l.allowAll(c) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code": http.StatusTooManyRequests,
				"msg":  "rate limit exceeded",
			})
			return
		}
		c.Next()
	}
}
