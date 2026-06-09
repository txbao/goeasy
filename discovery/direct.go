package discovery

import (
	"context"
	"fmt"
)

// directRegistry 静态服务表（config.discovery.services）。
type directRegistry struct {
	services map[string]string
}

func newDirectRegistry(services map[string]string) Registry {
	cp := make(map[string]string, len(services))
	for k, v := range services {
		cp[k] = v
	}
	return &directRegistry{services: cp}
}

func (d *directRegistry) Register(ctx context.Context, service, addr string) error {
	_ = ctx
	if service == "" || addr == "" {
		return fmt.Errorf("discovery: service and addr required")
	}
	d.services[service] = addr
	return nil
}

func (d *directRegistry) Resolve(ctx context.Context, service string) ([]string, error) {
	_ = ctx
	if addr, ok := d.services[service]; ok && addr != "" {
		return []string{addr}, nil
	}
	return nil, fmt.Errorf("discovery: service %q not in direct map", service)
}

func (d *directRegistry) Deregister(ctx context.Context) error {
	_ = ctx
	return nil
}
