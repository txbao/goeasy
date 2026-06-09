package discovery

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/txbao/goeasy/config"
)

// Registry 服务注册发现抽象。
type Registry interface {
	Register(ctx context.Context, service string, addr string) error
	Resolve(ctx context.Context, service string) ([]string, error)
	Deregister(ctx context.Context) error
}

type noopRegistry struct{}

func (n *noopRegistry) Register(ctx context.Context, service, addr string) error {
	_ = ctx
	_ = service
	_ = addr
	return nil
}

func (n *noopRegistry) Resolve(ctx context.Context, service string) ([]string, error) {
	_ = ctx
	_ = service
	return nil, errors.New("discovery: noop")
}

func (n *noopRegistry) Deregister(ctx context.Context) error {
	_ = ctx
	return nil
}

// NewRegistry 按配置创建服务发现（direct 默认；mode=etcd 且 enabled 时用 etcd，失败回退 direct）。
func NewRegistry(cfg *config.Config) Registry {
	if cfg == nil {
		return &noopRegistry{}
	}
	direct := newDirectRegistry(cfg.Discovery.Services)
	mode := strings.ToLower(strings.TrimSpace(cfg.Discovery.Mode))
	if mode == "etcd" && cfg.Discovery.Etcd.Enabled {
		reg, err := newEtcdRegistry(cfg.Discovery.Etcd, direct)
		if err == nil {
			return reg
		}
		log.Printf("discovery: etcd init failed, fallback to direct: %v", err)
	}
	return direct
}
