package scheduler

import "context"

// Scheduler 定时任务抽象。
type Scheduler interface {
	Register(name, spec string, fn func(context.Context) error) error
	Start() error
	Stop() error
}

type noopScheduler struct{}

func NewScheduler(cfg any) Scheduler {
	_ = cfg
	return &noopScheduler{}
}

func (n *noopScheduler) Register(name, spec string, fn func(context.Context) error) error {
	_ = name
	_ = spec
	_ = fn
	return nil
}

func (n *noopScheduler) Start() error { return nil }
func (n *noopScheduler) Stop() error  { return nil }
