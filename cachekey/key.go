package cachekey

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/txbao/goeasy/cache"
)

// EntityKey 实体缓存 key：{prefix}:{module}:id:{id}（module 为逻辑模块名，如 sys_roles）。
func EntityKey(prefix, module, id string) string {
	return fmt.Sprintf("%s:%s:id:%s", strings.TrimSpace(prefix), strings.TrimSpace(module), strings.TrimSpace(id))
}

// DeleteEntity 删除单条实体缓存（Update/Delete 后调用）。
func DeleteEntity(ctx context.Context, c cache.Cache, prefix, module, id string) error {
	if c == nil {
		return nil
	}
	return c.Del(ctx, EntityKey(prefix, module, id))
}

// SetEntityBytes 写入实体缓存。
func SetEntityBytes(ctx context.Context, c cache.Cache, prefix, module, id string, value []byte, ttl time.Duration) error {
	return c.SetBytes(ctx, EntityKey(prefix, module, id), value, ttl)
}

// GetEntityBytes 读取实体缓存；未命中返回 cache.ErrNotFound。
func GetEntityBytes(ctx context.Context, c cache.Cache, prefix, module, id string) ([]byte, error) {
	return c.GetBytes(ctx, EntityKey(prefix, module, id))
}
