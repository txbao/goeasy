package cache

import (
	"context"
	"errors"
	"time"

	"github.com/txbao/goeasy/config"
)

// Cache Redis 抽象。
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Close() error
}

// Locker 分布式锁占位。
type Locker interface {
	Lock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
}

func Open(cfg config.Redis) (Cache, error) {
	if !cfg.Enabled {
		return NewNoop(), nil
	}
	if cfg.Addr == "" {
		return nil, errors.New("redis enabled but addr is empty")
	}
	return &redisCache{addr: cfg.Addr}, nil
}

type redisCache struct{ addr string }

func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	_ = ctx
	_ = key
	return "", errors.New("redis: not connected (stub)")
}

func (r *redisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	_ = ctx
	_ = key
	_ = value
	_ = ttl
	return nil
}

func (r *redisCache) Close() error { return nil }

type noopCache struct{}

func NewNoop() Cache { return &noopCache{} }

func (n *noopCache) Get(ctx context.Context, key string) (string, error) {
	_ = ctx
	_ = key
	return "", errors.New("cache: noop")
}
func (n *noopCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	_ = ctx
	_ = key
	_ = value
	_ = ttl
	return nil
}
func (n *noopCache) Close() error { return nil }
