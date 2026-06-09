package eventbus

import (
	"context"
	"sync"
)

// Event 进程内事件。
type Event struct {
	Name    string
	Payload any
}

// Handler 同步事件处理器。
type Handler func(ctx context.Context, evt Event) error

// Bus 进程内事件总线（域事件同步分发；跨服务请用 mq.Envelope）。
type Bus interface {
	Publish(ctx context.Context, evt Event) error
	Subscribe(name string, h Handler)
}

type inProcessBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

// New 创建进程内事件总线。
func New() Bus {
	return &inProcessBus{handlers: make(map[string][]Handler)}
}

func (b *inProcessBus) Subscribe(name string, h Handler) {
	if b == nil || h == nil || name == "" {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[name] = append(b.handlers[name], h)
}

func (b *inProcessBus) Publish(ctx context.Context, evt Event) error {
	if b == nil {
		return nil
	}
	b.mu.RLock()
	hs := append([]Handler(nil), b.handlers[evt.Name]...)
	b.mu.RUnlock()
	for _, h := range hs {
		if err := h(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}
