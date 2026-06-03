package grpcx

import (
	"context"
	"fmt"

	"github.com/txbao/goeasy/breaker"
	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/loadbalance"
	"github.com/txbao/goeasy/retry"
)

// Client 带治理能力的 gRPC 客户端封装（调用面占位，可接真实 grpc.Dial）。
type Client struct {
	target     string
	timeoutSec int
	maxRetries int
	breaker    *breaker.Breaker
	retryCfg   config.RetryCfg
	balancer   loadbalance.Balancer
}

type ClientOption struct {
	Target     string
	TimeoutSec int
	MaxRetries int
	Cfg        *config.Config
	Strategy   string
}

func NewClient(opt ClientOption) (*Client, error) {
	if opt.Target == "" {
		return nil, fmt.Errorf("grpc client target is empty")
	}
	var cfg *config.Config
	if opt.Cfg != nil {
		cfg = opt.Cfg
	}
	c := &Client{
		target:     opt.Target,
		timeoutSec: opt.TimeoutSec,
		maxRetries: opt.MaxRetries,
		balancer:   loadbalance.New(opt.Strategy),
	}
	if cfg != nil {
		c.breaker = breaker.New(cfg.Governance.Breaker)
		c.retryCfg = cfg.Governance.Retry
	}
	return c, nil
}

// Invoke 执行远程调用（占位：集成熔断与重试）。
func (c *Client) Invoke(ctx context.Context, fn func(ctx context.Context) error) error {
	run := func() error {
		if c.breaker != nil {
			_, err := c.breaker.Execute(func() (any, error) {
				return nil, fn(ctx)
			})
			return err
		}
		return fn(ctx)
	}
	return retry.Do(ctx, c.retryCfg, run)
}

func (c *Client) Target() string { return c.target }
