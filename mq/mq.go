package mq

import (
	"context"
	"errors"
	"strings"

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
	typ := strings.ToLower(strings.TrimSpace(cfg.Type))
	if typ == "" {
		typ = "nsq"
	}
	switch typ {
	case "nsq":
		return openNSQ(cfg)
	default:
		return nil, errors.New("mq: unsupported type " + typ)
	}
}

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
