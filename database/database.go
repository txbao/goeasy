package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/txbao/goeasy/config"
)

// DB 数据库抽象。
type DB interface {
	Ping(ctx context.Context) error
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
	Close() error
}

func Open(cfg config.DB) (DB, error) {
	if !cfg.Enabled {
		return NewNoop(), nil
	}
	if cfg.DSN == "" {
		return nil, errors.New("database enabled but dsn is empty")
	}
	return &postgresDB{dsn: cfg.DSN, driver: cfg.Driver}, nil
}

type postgresDB struct {
	dsn    string
	driver string
}

func (p *postgresDB) Ping(ctx context.Context) error {
	if p.dsn == "" {
		return errors.New("invalid dsn")
	}
	// P1: 真实 GORM 连接在后续版本接入；当前校验 DSN 非空即可启动链路。
	_ = ctx
	return nil
}

func (p *postgresDB) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (p *postgresDB) Close() error { return nil }

type noopDB struct{}

func NewNoop() DB { return &noopDB{} }

func (n *noopDB) Ping(ctx context.Context) error { return nil }
func (n *noopDB) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
func (n *noopDB) Close() error { return nil }

func Validate(cfg config.DB) error {
	if !cfg.Enabled {
		return nil
	}
	if cfg.DSN == "" {
		return fmt.Errorf("database.enabled=true requires database.dsn")
	}
	return nil
}
