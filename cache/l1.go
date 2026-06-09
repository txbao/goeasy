package cache

import (
	"context"
	"sync"
	"time"
)

type l1Entry struct {
	value     []byte
	expiresAt time.Time
}

// L1Cache 进程内 LRU 近似缓存（按 maxEntries 淘汰）。
type L1Cache struct {
	mu         sync.RWMutex
	data       map[string]l1Entry
	maxEntries int
	defaultTTL time.Duration
}

func NewL1(maxEntries int, defaultTTL time.Duration) *L1Cache {
	if maxEntries <= 0 {
		maxEntries = 10000
	}
	if defaultTTL <= 0 {
		defaultTTL = time.Minute
	}
	return &L1Cache{data: make(map[string]l1Entry), maxEntries: maxEntries, defaultTTL: defaultTTL}
}

func (l *L1Cache) get(key string) ([]byte, bool) {
	l.mu.RLock()
	e, ok := l.data[key]
	l.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

func (l *L1Cache) set(key string, value []byte, ttl time.Duration) {
	if ttl <= 0 {
		ttl = l.defaultTTL
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.data) >= l.maxEntries {
		for k := range l.data {
			delete(l.data, k)
			break
		}
	}
	l.data[key] = l1Entry{value: append([]byte(nil), value...), expiresAt: time.Now().Add(ttl)}
}

func (l *L1Cache) del(keys ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, k := range keys {
		delete(l.data, k)
	}
}

// MultiLevelCache L1 内存 + L2 Redis。
type MultiLevelCache struct {
	l1 *L1Cache
	l2 Cache
}

func NewMultiLevel(l1 *L1Cache, l2 Cache) Cache {
	if l1 == nil {
		return l2
	}
	if l2 == nil {
		return NewNoop()
	}
	return &MultiLevelCache{l1: l1, l2: l2}
}

func (m *MultiLevelCache) Get(ctx context.Context, key string) (string, error) {
	b, err := m.GetBytes(ctx, key)
	return string(b), err
}

func (m *MultiLevelCache) GetBytes(ctx context.Context, key string) ([]byte, error) {
	if m.l1 != nil {
		if v, ok := m.l1.get(key); ok {
			if isNullMarker(v) {
				return nil, ErrNotFound
			}
			return v, nil
		}
	}
	val, err := m.l2.GetBytes(ctx, key)
	if err != nil {
		return nil, err
	}
	if m.l1 != nil {
		m.l1.set(key, val, m.l1.defaultTTL)
	}
	return val, nil
}

func (m *MultiLevelCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return m.SetBytes(ctx, key, []byte(value), ttl)
}

func (m *MultiLevelCache) SetBytes(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.l1 != nil {
		m.l1.set(key, value, m.l1.defaultTTL)
	}
	return m.l2.SetBytes(ctx, key, value, ttl)
}

func (m *MultiLevelCache) Del(ctx context.Context, keys ...string) error {
	if m.l1 != nil {
		m.l1.del(keys...)
	}
	return m.l2.Del(ctx, keys...)
}

func (m *MultiLevelCache) Close() error { return m.l2.Close() }
