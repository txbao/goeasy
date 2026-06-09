package grpcx

import (
	"context"
	"testing"
	"time"

	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/discovery"
)

func TestRegistryReuseDirect(t *testing.T) {
	cfg := &config.Config{
		GRPC: config.GRPC{TimeoutSec: 1, MaxRetries: 0},
		Discovery: config.Discovery{
			Mode: "direct",
			Services: map[string]string{
				"peer": "127.0.0.1:1", // 不可达；仅测 Resolve + 缓存路径需 mock
			},
		},
	}
	reg := NewRegistry(cfg, discovery.NewRegistry(cfg))

	// 第一次 dial 会失败（端口不可达），但可验证空 service 等边界
	_, err := reg.Client(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty service")
	}

	target, err := ResolveService(context.Background(), cfg, discovery.NewRegistry(cfg), "peer")
	if err != nil || target != "127.0.0.1:1" {
		t.Fatalf("resolve: got %q err=%v", target, err)
	}
}

func TestRegistryClientLazyNonBlocking(t *testing.T) {
	cfg := &config.Config{
		GRPC: config.GRPC{TimeoutSec: 1, MaxRetries: 0},
		Discovery: config.Discovery{
			Mode:     "direct",
			Services: map[string]string{"peer": "127.0.0.1:1"},
		},
	}
	reg := NewRegistry(cfg, discovery.NewRegistry(cfg))
	done := make(chan error, 1)
	go func() {
		_, err := reg.ClientLazy(context.Background(), "peer")
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("ClientLazy should not fail on dial: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ClientLazy blocked too long; expected non-blocking dial")
	}
}

func TestRegistryCloseNil(t *testing.T) {
	var reg *Registry
	if err := reg.Close(); err != nil {
		t.Fatal(err)
	}
}
