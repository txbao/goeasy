package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNoopCacheNotFound(t *testing.T) {
	c := NewNoop()
	ctx := context.Background()
	_, err := c.GetBytes(ctx, "k")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := c.Del(ctx, "k"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetBytes(ctx, "k", []byte("v"), time.Hour); err != nil {
		t.Fatal(err)
	}
}
