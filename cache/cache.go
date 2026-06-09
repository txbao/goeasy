package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/txbao/goeasy/config"
)

// ErrNotFound 缓存未命中（与 Redis Nil 区分，供仓储 cache-aside）。
var ErrNotFound = errors.New("cache: key not found")

// Cache Redis 抽象。
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	GetBytes(ctx context.Context, key string) ([]byte, error)
	SetBytes(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
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
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &redisCache{client: client}, nil
}

type redisCache struct {
	client *redis.Client
}

func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	return val, err
}

func (r *redisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrNotFound
	}
	return val, err
}

func (r *redisCache) SetBytes(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisCache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *redisCache) Close() error { return r.client.Close() }

type noopCache struct{}

func NewNoop() Cache { return &noopCache{} }

func (n *noopCache) Get(ctx context.Context, key string) (string, error) {
	_ = ctx
	_ = key
	return "", ErrNotFound
}
func (n *noopCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	_ = ctx
	_ = key
	_ = value
	_ = ttl
	return nil
}
func (n *noopCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	_ = ctx
	_ = key
	return nil, ErrNotFound
}
func (n *noopCache) SetBytes(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	_ = ctx
	_ = key
	_ = value
	_ = ttl
	return nil
}
func (n *noopCache) Del(ctx context.Context, keys ...string) error {
	_ = ctx
	_ = keys
	return nil
}
func (n *noopCache) Close() error { return nil }
