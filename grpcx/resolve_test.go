package grpcx

import (
	"context"
	"testing"

	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/discovery"
)

func TestResolveServiceYamlOverridesEtcd(t *testing.T) {
	cfg := &config.Config{
		Discovery: config.Discovery{
			Mode: "etcd",
			Etcd: config.EtcdDiscoveryCfg{Enabled: true, Endpoints: []string{"127.0.0.1:2379"}},
			Services: map[string]string{
				"demo1": "127.0.0.1:18083",
			},
		},
	}
	target, err := ResolveService(context.Background(), cfg, discovery.NewRegistry(cfg), "demo1")
	if err != nil || target != "127.0.0.1:18083" {
		t.Fatalf("yaml services should win over etcd: got %q err=%v", target, err)
	}
}

func TestResolveServiceDirect(t *testing.T) {
	cfg := &config.Config{
		Discovery: config.Discovery{
			Mode: "direct",
			Services: map[string]string{
				"peer": "localhost:9001",
			},
		},
	}
	target, err := ResolveService(context.Background(), cfg, discovery.NewRegistry(cfg), "peer")
	if err != nil || target != "localhost:9001" {
		t.Fatalf("got %q err=%v", target, err)
	}
}
