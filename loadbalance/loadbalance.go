package loadbalance

import (
	"math/rand"
	"sync/atomic"
)

// Balancer 负载均衡策略。
type Balancer interface {
	Next(addrs []string) string
}

type RoundRobin struct{ n uint64 }

func (r *RoundRobin) Next(addrs []string) string {
	if len(addrs) == 0 {
		return ""
	}
	i := atomic.AddUint64(&r.n, 1)
	return addrs[int(i-1)%len(addrs)]
}

type Random struct{}

func (r *Random) Next(addrs []string) string {
	if len(addrs) == 0 {
		return ""
	}
	return addrs[rand.Intn(len(addrs))]
}

func New(strategy string) Balancer {
	switch strategy {
	case "random":
		return &Random{}
	default:
		return &RoundRobin{}
	}
}
