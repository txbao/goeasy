package discovery

import (
	"context"
	"errors"
)

// Registry 服务注册发现抽象。
type Registry interface {
	Register(ctx context.Context, service string, addr string) error
	Resolve(ctx context.Context, service string) ([]string, error)
}

type noopRegistry struct{}

func NewRegistry(cfg any) Registry {
	_ = cfg
	return &noopRegistry{}
}

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
