package grpcx

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/txbao/goeasy/breaker"
	"github.com/txbao/goeasy/config"
	"github.com/txbao/goeasy/loadbalance"
	"github.com/txbao/goeasy/retry"
)

// Client 带治理能力的 gRPC 客户端。
type Client struct {
	target     string
	timeoutSec int
	maxRetries int
	breaker    *breaker.Breaker
	retryCfg   config.RetryCfg
	balancer   loadbalance.Balancer
	conn       *grpc.ClientConn
}

type ClientOption struct {
	Target     string
	TimeoutSec int
	MaxRetries int
	Cfg        *config.Config
	Strategy   string
	// Lazy 为 true 时不阻塞 dial（首次 RPC 再建连）；bootstrap 注册 HTTP 路由时使用，避免对端未启动导致进程无法监听端口。
	Lazy bool
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
	if c.timeoutSec <= 0 {
		c.timeoutSec = 5
	}
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	var conn *grpc.ClientConn
	var err error
	if opt.Lazy {
		conn, err = grpc.DialContext(context.Background(), opt.Target, dialOpts...)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.timeoutSec)*time.Second)
		defer cancel()
		conn, err = grpc.DialContext(ctx, opt.Target, append(dialOpts, grpc.WithBlock())...)
	}
	if err != nil {
		return nil, fmt.Errorf("grpc dial %s: %w", opt.Target, err)
	}
	c.conn = conn
	return c, nil
}

// Conn 返回底层连接（供生成的 pb 客户端使用）。
func (c *Client) Conn() *grpc.ClientConn {
	if c == nil {
		return nil
	}
	return c.conn
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Invoke 执行远程调用（熔断 + 重试包装）。
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

func (c *Client) Target() string {
	if c == nil {
		return ""
	}
	return c.target
}
