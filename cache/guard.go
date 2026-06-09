package cache

import (
	"context"
	"sync"
	"time"
)

var nullMarker = []byte("__goeasy_null__")

func isNullMarker(b []byte) bool {
	return string(b) == string(nullMarker)
}

// NullBytes 空值缓存标记（防穿透）。
func NullBytes() []byte {
	return append([]byte(nil), nullMarker...)
}

// JitterTTL 在 base TTL 上增加随机抖动（防雪崩）。
func JitterTTL(base time.Duration, jitterRatio float64) time.Duration {
	if base <= 0 || jitterRatio <= 0 {
		return base
	}
	if jitterRatio > 1 {
		jitterRatio = 1
	}
	delta := time.Duration(float64(base) * jitterRatio)
	// 简单伪随机：纳秒低位
	n := time.Now().UnixNano() % int64(delta*2+1)
	return base - delta/2 + time.Duration(n)
}

// GuardedGet 带 singleflight 的缓存读取，miss 时调用 loader 回填（防击穿）。
type GuardedGet struct {
	mu sync.Mutex
	in map[string]*call
}

type call struct {
	wg  sync.WaitGroup
	val []byte
	err error
}

func NewGuardedGet() *GuardedGet {
	return &GuardedGet{in: make(map[string]*call)}
}

// Do 读取缓存；miss 时仅一次调用 loader 并写入 cache。
func (g *GuardedGet) Do(ctx context.Context, c Cache, key string, ttl time.Duration, loader func(context.Context) ([]byte, error)) ([]byte, error) {
	if c != nil {
		if raw, err := c.GetBytes(ctx, key); err == nil {
			return raw, nil
		} else if err != ErrNotFound {
			return nil, err
		}
	}
	g.mu.Lock()
	if g.in == nil {
		g.in = make(map[string]*call)
	}
	if call, ok := g.in[key]; ok {
		g.mu.Unlock()
		call.wg.Wait()
		return call.val, call.err
	}
	call := &call{}
	call.wg.Add(1)
	g.in[key] = call
	g.mu.Unlock()

	call.val, call.err = loader(ctx)
	if call.err == nil && c != nil {
		_ = c.SetBytes(ctx, key, call.val, ttl)
	}

	g.mu.Lock()
	delete(g.in, key)
	g.mu.Unlock()
	call.wg.Done()
	return call.val, call.err
}

// SetNull 写入空值缓存（防穿透）。
func SetNull(ctx context.Context, c Cache, key string, ttl time.Duration) error {
	if c == nil {
		return nil
	}
	return c.SetBytes(ctx, key, NullBytes(), ttl)
}
