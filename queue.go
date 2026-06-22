// Package queue is togo's default in-process job queue provider. Blank-import
// (or `togo install togo-framework/queue`) to register it with the kernel.
package queue

import (
	"context"
	"sync"

	"github.com/togo-framework/togo"
	tqueue "github.com/togo-framework/togo/queue"
)

func init() {
	togo.RegisterProviderFunc("queue", togo.PriorityLate, func(k *togo.Kernel) error {
		k.Queue = NewMemory(func(err error) { k.Log.Error("queue job failed", "err", err) })
		return nil
	})
}

type memory struct {
	mu       sync.RWMutex
	handlers map[string]tqueue.Handler
	onError  func(error)
}

// NewMemory returns an in-process queue.
func NewMemory(onError func(error)) tqueue.Queue {
	return &memory{handlers: map[string]tqueue.Handler{}, onError: onError}
}

func (m *memory) Handle(name string, h tqueue.Handler) {
	m.mu.Lock()
	m.handlers[name] = h
	m.mu.Unlock()
}

func (m *memory) Dispatch(ctx context.Context, name string, payload any) error {
	m.mu.RLock()
	h, ok := m.handlers[name]
	m.mu.RUnlock()
	if !ok {
		return nil
	}
	go func() {
		if err := h(context.WithoutCancel(ctx), payload); err != nil && m.onError != nil {
			m.onError(err)
		}
	}()
	return nil
}
