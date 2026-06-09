package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisLimiter 基于 Redis 的固定窗口计数限流。
type redisLimiter struct {
	client *redis.Client
	qps    float64
	burst  int
	prefix string
}

func (r *redisLimiter) allow(key string) bool {
	if r == nil || r.client == nil {
		return true
	}
	if key == "" {
		key = "global"
	}
	limit := r.burst
	if limit <= 0 {
		limit = int(r.qps)
	}
	if limit <= 0 {
		limit = 100
	}
	window := time.Second
	if r.qps > 0 && r.qps < 1 {
		window = time.Duration(float64(time.Second) / r.qps)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	redisKey := fmt.Sprintf("%s:%s:%d", r.prefix, key, time.Now().Unix()/int64(window.Seconds()))
	if window < time.Second {
		redisKey = fmt.Sprintf("%s:%s:%d", r.prefix, key, time.Now().UnixNano()/int64(window))
	}
	n, err := r.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return true
	}
	if n == 1 {
		_ = r.client.Expire(ctx, redisKey, window*2).Err()
	}
	return int(n) <= limit
}
