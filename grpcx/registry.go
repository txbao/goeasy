package grpcx

import (
	"context"
	"fmt"
	"sync"

	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/discovery"
)

// Registry 按逻辑服务名缓存 gRPC 长连接（类似 go-zero zrpc 客户端池）。
type Registry struct {
	cfg        *config.Config
	reg        discovery.Registry
	timeoutSec int
	maxRetries int
	mu         sync.Mutex
	clients    map[string]*Client
}

// NewRegistry 创建 RPC 客户端注册表；reg 通常为 discovery.NewRegistry(cfg)。
func NewRegistry(cfg *config.Config, reg discovery.Registry) *Registry {
	if cfg == nil {
		cfg = &config.Config{}
	}
	return &Registry{
		cfg:        cfg,
		reg:        reg,
		timeoutSec: cfg.GRPC.TimeoutSec,
		maxRetries: cfg.GRPC.MaxRetries,
		clients:    make(map[string]*Client),
	}
}

// Client 返回 service 对应的长连接客户端；同一 service 进程内复用连接。
func (r *Registry) Client(ctx context.Context, service string) (*Client, error) {
	if r == nil {
		return nil, fmt.Errorf("grpc registry is nil")
	}
	r.mu.Lock()
	if c, ok := r.clients[service]; ok {
		r.mu.Unlock()
		return c, nil
	}
	r.mu.Unlock()

	target, err := ResolveService(ctx, r.cfg, r.reg, service)
	if err != nil {
		return nil, err
	}
	cli, err := NewClient(ClientOption{
		Target:     target,
		TimeoutSec: r.timeoutSec,
		MaxRetries: r.maxRetries,
		Cfg:        r.cfg,
		Strategy:   "round_robin",
	})
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.clients[service]; ok {
		_ = cli.Close()
		return c, nil
	}
	r.clients[service] = cli
	return cli, nil
}

// ClientLazy 同 Client，但不阻塞 dial；HTTP 启动阶段装配 Gateway 时使用，对端未就绪时仍可监听业务端口。
func (r *Registry) ClientLazy(ctx context.Context, service string) (*Client, error) {
	if r == nil {
		return nil, fmt.Errorf("grpc registry is nil")
	}
	r.mu.Lock()
	if c, ok := r.clients[service]; ok {
		r.mu.Unlock()
		return c, nil
	}
	r.mu.Unlock()

	target, err := ResolveService(ctx, r.cfg, r.reg, service)
	if err != nil {
		return nil, err
	}
	cli, err := NewClient(ClientOption{
		Target:     target,
		TimeoutSec: r.timeoutSec,
		MaxRetries: r.maxRetries,
		Cfg:        r.cfg,
		Strategy:   "round_robin",
		Lazy:       true,
	})
	if err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.clients[service]; ok {
		_ = cli.Close()
		return c, nil
	}
	r.clients[service] = cli
	return cli, nil
}

// MustClient 同 Client，失败 panic（仅 bootstrap 使用）。
func (r *Registry) MustClient(ctx context.Context, service string) *Client {
	cli, err := r.Client(ctx, service)
	if err != nil {
		panic(fmt.Sprintf("grpc registry client %q: %v", service, err))
	}
	return cli
}

// Close 关闭全部缓存连接；进程退出前调用。
func (r *Registry) Close() error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	var first error
	for name, c := range r.clients {
		if err := c.Close(); err != nil && first == nil {
			first = fmt.Errorf("close grpc client %q: %w", name, err)
		}
		delete(r.clients, name)
	}
	return first
}
