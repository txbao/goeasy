package cache

import "github.com/redis/go-redis/v9"

// RedisClient 从 Cache 提取底层 Redis 客户端（供限流/锁等复用）。
func RedisClient(c Cache) *redis.Client {
	if c == nil {
		return nil
	}
	if m, ok := c.(*MultiLevelCache); ok && m.l2 != nil {
		return RedisClient(m.l2)
	}
	if r, ok := c.(*redisCache); ok {
		return r.client
	}
	return nil
}
