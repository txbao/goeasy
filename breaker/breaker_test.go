package breaker

import (
	"testing"

	"github.com/txbao/goeasy/config"
)

func TestExecuteDisabled(t *testing.T) {
	b := New(config.BreakerCfg{Enabled: false})
	v, err := b.Execute(func() (any, error) {
		return "ok", nil
	})
	if err != nil || v != "ok" {
		t.Fatal(err)
	}
}

func TestExecuteEnabled(t *testing.T) {
	b := New(config.BreakerCfg{Enabled: true, MaxRequests: 2, IntervalSec: 60, TimeoutSec: 60})
	if b.State() == "disabled" {
		t.Fatal("expected active breaker")
	}
}
