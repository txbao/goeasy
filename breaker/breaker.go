package breaker

import (
	"time"

	"github.com/sony/gobreaker"

	"github.com/txbao/goeasy/config"
)

// Breaker 熔断器封装。
type Breaker struct {
	cb *gobreaker.CircuitBreaker
}

func New(cfg config.BreakerCfg) *Breaker {
	if !cfg.Enabled {
		return &Breaker{cb: nil}
	}
	st := gobreaker.Settings{
		Name:        "goesy",
		MaxRequests: cfg.MaxRequests,
		Interval:    time.Duration(cfg.IntervalSec) * time.Second,
		Timeout:     time.Duration(cfg.TimeoutSec) * time.Second,
	}
	return &Breaker{cb: gobreaker.NewCircuitBreaker(st)}
}

func (b *Breaker) Execute(fn func() (any, error)) (any, error) {
	if b == nil || b.cb == nil {
		return fn()
	}
	return b.cb.Execute(fn)
}

func (b *Breaker) State() string {
	if b == nil || b.cb == nil {
		return "disabled"
	}
	return b.cb.State().String()
}
