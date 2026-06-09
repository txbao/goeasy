package scheduler

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/txbao/goeasy/config"
)

// TaskKind 任务分类。
type TaskKind string

const (
	TaskSystem   TaskKind = "system"
	TaskBusiness TaskKind = "business"
)

// Scheduler 定时任务抽象。
type Scheduler interface {
	Register(name, spec string, fn func(context.Context) error) error
	RegisterKind(kind TaskKind, name, spec string, fn func(context.Context) error) error
	Start() error
	Stop() error
}

type cronScheduler struct {
	cron   *cron.Cron
	mu     sync.Mutex
	started bool
}

func NewScheduler(cfg config.SchedulerCfg) Scheduler {
	if !cfg.Enabled {
		return &noopScheduler{}
	}
	loc := time.Local
	if tz := strings.TrimSpace(cfg.Timezone); tz != "" {
		if l, err := time.LoadLocation(tz); err == nil {
			loc = l
		}
	}
	return &cronScheduler{cron: cron.New(cron.WithLocation(loc), cron.WithSeconds())}
}

// New 兼容旧签名。
func New(cfg any) Scheduler {
	if c, ok := cfg.(*config.Config); ok && c != nil {
		return NewScheduler(c.Scheduler)
	}
	if c, ok := cfg.(config.SchedulerCfg); ok {
		return NewScheduler(c)
	}
	return &noopScheduler{}
}

func (s *cronScheduler) Register(name, spec string, fn func(context.Context) error) error {
	return s.RegisterKind(TaskBusiness, name, spec, fn)
}

func (s *cronScheduler) RegisterKind(kind TaskKind, name, spec string, fn func(context.Context) error) error {
	if s == nil || s.cron == nil || fn == nil {
		return errors.New("scheduler: invalid register")
	}
	if name == "" || spec == "" {
		return errors.New("scheduler: name and spec required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.cron.AddFunc(spec, func() {
		ctx := context.WithValue(context.Background(), taskKindKey{}, kind)
		ctx = context.WithValue(ctx, taskNameKey{}, name)
		_ = fn(ctx)
	})
	return err
}

type taskKindKey struct{}
type taskNameKey struct{}

// TaskKindFrom 从 context 读取任务分类。
func TaskKindFrom(ctx context.Context) TaskKind {
	if v, ok := ctx.Value(taskKindKey{}).(TaskKind); ok {
		return v
	}
	return TaskBusiness
}

// TaskNameFrom 从 context 读取任务名。
func TaskNameFrom(ctx context.Context) string {
	if v, ok := ctx.Value(taskNameKey{}).(string); ok {
		return v
	}
	return ""
}

func (s *cronScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.started {
		return nil
	}
	s.cron.Start()
	s.started = true
	return nil
}

func (s *cronScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started || s.cron == nil {
		return nil
	}
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.started = false
	return nil
}

type noopScheduler struct{}

func (n *noopScheduler) Register(name, spec string, fn func(context.Context) error) error {
	_ = name
	_ = spec
	_ = fn
	return nil
}

func (n *noopScheduler) RegisterKind(kind TaskKind, name, spec string, fn func(context.Context) error) error {
	_ = kind
	return n.Register(name, spec, fn)
}

func (n *noopScheduler) Start() error { return nil }
func (n *noopScheduler) Stop() error  { return nil }
