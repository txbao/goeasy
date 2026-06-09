package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("app_name: test\nhttp:\n  port: 9090\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AppName != "test" || cfg.HTTP.Port != 9090 {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
}

func TestRedisKeyPrefixAndEntityCache(t *testing.T) {
	cfg := &Config{
		AppName: "demo",
		Redis:   Redis{Enabled: true, KeyPrefix: "app"},
		Cache:   CacheCfg{Enabled: true, EntityTTL: "24h"},
	}
	applyDefaults(cfg)
	if cfg.RedisKeyPrefix() != "app" {
		t.Fatalf("prefix: %s", cfg.RedisKeyPrefix())
	}
	if cfg.EntityCacheTTL() != 24*time.Hour {
		t.Fatalf("ttl: %v", cfg.EntityCacheTTL())
	}
	if !cfg.EntityCacheEnabled() {
		t.Fatal("expected entity cache enabled")
	}
	cfg.Cache.Enabled = false
	if cfg.EntityCacheEnabled() {
		t.Fatal("expected disabled when cache.enabled false")
	}
	cfg2 := &Config{AppName: "fallback"}
	applyDefaults(cfg2)
	if cfg2.RedisKeyPrefix() != "fallback" {
		t.Fatalf("fallback prefix: %s", cfg2.RedisKeyPrefix())
	}
}
