package retry

import (
	"context"
	"time"

	"github.com/txbao/goeasy/config"
)

// Do 按配置重试。
func Do(ctx context.Context, cfg config.RetryCfg, fn func() error) error {
	if !cfg.Enabled {
		return fn()
	}
	var err error
	backoff := time.Duration(cfg.BackoffMs) * time.Millisecond
	for i := 0; i <= cfg.MaxRetries; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err = fn(); err == nil {
			return nil
		}
		if i < cfg.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
	return err
}
