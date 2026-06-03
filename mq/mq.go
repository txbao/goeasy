package mq

import (
	"context"
	"errors"

	"github.com/txbao/goeasy/config"
)

// MQ 消息队列抽象。
type MQ interface {
	Publish(ctx context.Context, topic string, body []byte) error
	Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, body []byte) error) error
	Close() error
}

func Open(cfg config.MQ) (MQ, error) {
	if !cfg.Enabled {
		return NewNoop(), nil
	}
	if cfg.Addr == "" {
		return nil, errors.New("mq enabled but addr is empty")
	}
	return &nsqMQ{addr: cfg.Addr, typ: cfg.Type}, nil
}

type nsqMQ struct {
	addr string
	typ  string
}

func (n *nsqMQ) Publish(ctx context.Context, topic string, body []byte) error {
	_ = ctx
	_ = topic
	_ = body
	return nil
}

func (n *nsqMQ) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, body []byte) error) error {
	_ = ctx
	_ = topic
	_ = handler
	return nil
}

func (n *nsqMQ) Close() error { return nil }

type noopMQ struct{}

func NewNoop() MQ { return &noopMQ{} }

func (n *noopMQ) Publish(ctx context.Context, topic string, body []byte) error {
	_ = ctx
	_ = topic
	_ = body
	return nil
}
func (n *noopMQ) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, body []byte) error) error {
	_ = ctx
	_ = topic
	_ = handler
	return nil
}
func (n *noopMQ) Close() error { return nil }
