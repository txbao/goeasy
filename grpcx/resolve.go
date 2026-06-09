package grpcx

import (
	"context"
	"fmt"
	"strings"

	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/discovery"
	"github.com/txbao/goeasy/loadbalance"
)

// ResolveService 解析逻辑服务名为 gRPC target（discovery.services 显式配置优先，否则走 Registry/ETCD）。
func ResolveService(ctx context.Context, cfg *config.Config, reg discovery.Registry, service string) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("grpc: config is nil")
	}
	service = strings.TrimSpace(service)
	if service == "" {
		return "", fmt.Errorf("grpc: empty service name")
	}
	// discovery.services 显式配置优先（本地 dev 直连，避免 etcd 未就绪时阻塞启动）。
	if t, ok := cfg.Discovery.Services[service]; ok && strings.TrimSpace(t) != "" {
		return strings.TrimSpace(t), nil
	}
	if strings.EqualFold(cfg.Discovery.Mode, "direct") || !cfg.Discovery.Etcd.Enabled {
		return "", fmt.Errorf("grpc: no endpoint for service %q in discovery.services", service)
	}
	if reg == nil {
		return "", fmt.Errorf("grpc: registry unavailable for %q", service)
	}
	addrs, err := reg.Resolve(ctx, service)
	if err != nil {
		return "", err
	}
	t := loadbalance.New("round_robin").Next(addrs)
	if t == "" {
		return "", fmt.Errorf("grpc: no endpoint for service %q", service)
	}
	return t, nil
}
