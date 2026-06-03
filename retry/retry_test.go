package retry

import (
	"context"
	"errors"
	"testing"

	"github.com/txbao/goeasy/config"
)

func TestDoSuccess(t *testing.T) {
	n := 0
	err := Do(context.Background(), config.RetryCfg{Enabled: true, MaxRetries: 2, BackoffMs: 1}, func() error {
		n++
		if n < 2 {
			return errors.New("retry")
		}
		return nil
	})
	if err != nil || n != 2 {
		t.Fatalf("got n=%d err=%v", n, err)
	}
}
