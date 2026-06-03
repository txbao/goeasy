package storage

import (
	"context"
	"io"
)

// Store 对象存储抽象。
type Store interface {
	Put(ctx context.Context, key string, r io.Reader) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}

type noopStore struct{}

func NewStore(cfg any) Store {
	_ = cfg
	return &noopStore{}
}

func (n *noopStore) Put(ctx context.Context, key string, r io.Reader) error {
	_ = ctx
	_ = key
	_ = r
	return nil
}

func (n *noopStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	_ = ctx
	_ = key
	return nil, nil
}
