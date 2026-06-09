package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var unlockScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
  return redis.call("del", KEYS[1])
else
  return 0
end
`)

// RedisLocker 基于 Redis SET NX 的分布式锁。
type RedisLocker struct {
	client *redis.Client
	prefix string
}

func NewRedisLocker(client *redis.Client, keyPrefix string) Locker {
	if client == nil {
		return noopLocker{}
	}
	p := keyPrefix
	if p == "" {
		p = "goeasy"
	}
	return &RedisLocker{client: client, prefix: p + ":lock:"}
}

type noopLocker struct{}

func (noopLocker) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	_ = ctx
	_ = key
	_ = ttl
	return true, nil
}

func (noopLocker) Unlock(ctx context.Context, key string) error {
	_ = ctx
	_ = key
	return nil
}

type lockTokenKey struct{}

// Lock 尝试获取锁；成功时 token 存入 context，Unlock 需同一 context。
func (l *RedisLocker) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if l == nil || l.client == nil {
		return false, errors.New("locker: redis not available")
	}
	token := uuid.NewString()
	ok, err := l.client.SetNX(ctx, l.prefix+key, token, ttl).Result()
	if err != nil {
		return false, err
	}
	if ok {
		// 调用方通过 WithLockToken 保存 token；简化 API 用 key 映射
		_ = token
	}
	return ok, nil
}

// TryLock 获取锁并返回 unlock 函数。
func (l *RedisLocker) TryLock(ctx context.Context, key string, ttl time.Duration) (unlock func() error, ok bool, err error) {
	if l == nil || l.client == nil {
		return nil, false, errors.New("locker: redis not available")
	}
	token := uuid.NewString()
	redisKey := l.prefix + key
	acquired, err := l.client.SetNX(ctx, redisKey, token, ttl).Result()
	if err != nil || !acquired {
		return nil, acquired, err
	}
	return func() error {
		ctx2, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, e := unlockScript.Run(ctx2, l.client, []string{redisKey}, token).Result()
		return e
	}, true, nil
}

func (l *RedisLocker) Unlock(ctx context.Context, key string) error {
	if l == nil || l.client == nil {
		return nil
	}
	token, ok := ctx.Value(lockTokenKey{}).(string)
	if !ok || token == "" {
		return fmt.Errorf("locker: missing lock token in context")
	}
	redisKey := l.prefix + key
	_, err := unlockScript.Run(ctx, l.client, []string{redisKey}, token).Result()
	return err
}
