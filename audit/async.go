package audit

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/txbao/goeasy/contextx"
)

// MultiRecorder 同时写入多个 Recorder（如 JSON Logger + 业务 DB）。
type MultiRecorder struct {
	recorders []Recorder
}

func NewMultiRecorder(recorders ...Recorder) Recorder {
	filtered := make([]Recorder, 0, len(recorders))
	for _, r := range recorders {
		if r != nil {
			filtered = append(filtered, r)
		}
	}
	if len(filtered) == 0 {
		return NopRecorder{}
	}
	if len(filtered) == 1 {
		return filtered[0]
	}
	return &MultiRecorder{recorders: filtered}
}

func (m *MultiRecorder) Record(ctx context.Context, op contextx.OperatorContext, entry Entry) error {
	var first error
	for _, r := range m.recorders {
		if err := r.Record(ctx, op, entry); err != nil && first == nil {
			first = err
		}
	}
	return first
}

type asyncTask struct {
	ctx   context.Context
	op    contextx.OperatorContext
	entry Entry
}

// AsyncRecorder 异步写入业务 Recorder；写库失败仅 slog.Warn，不影响主流程。
type AsyncRecorder struct {
	inner  Recorder
	ch     chan asyncTask
	wg     sync.WaitGroup
	closed chan struct{}
	log    *slog.Logger
}

func NewAsyncRecorder(inner Recorder, bufferSize int) *AsyncRecorder {
	if inner == nil {
		inner = NopRecorder{}
	}
	if bufferSize <= 0 {
		bufferSize = 256
	}
	a := &AsyncRecorder{
		inner:  inner,
		ch:     make(chan asyncTask, bufferSize),
		closed: make(chan struct{}),
		log:    slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
	a.wg.Add(1)
	go a.loop()
	return a
}

func (a *AsyncRecorder) loop() {
	defer a.wg.Done()
	for task := range a.ch {
		if err := a.inner.Record(task.ctx, task.op, task.entry); err != nil {
			a.log.Warn("audit async record failed", "error", err.Error(),
				"module", task.entry.ModuleCode, "action", task.entry.ActionType)
		}
	}
}

func (a *AsyncRecorder) Record(ctx context.Context, op contextx.OperatorContext, entry Entry) error {
	if a == nil || a.inner == nil {
		return nil
	}
	select {
	case <-a.closed:
		return a.inner.Record(ctx, op, entry)
	default:
	}
	task := asyncTask{ctx: ctx, op: op, entry: entry}
	select {
	case a.ch <- task:
	default:
		a.log.Warn("audit async buffer full, dropping entry",
			"module", entry.ModuleCode, "action", entry.ActionType)
	}
	return nil
}

// Close 优雅排空异步队列。
func (a *AsyncRecorder) Close() {
	if a == nil {
		return
	}
	select {
	case <-a.closed:
		return
	default:
		close(a.closed)
		close(a.ch)
		a.wg.Wait()
	}
}
