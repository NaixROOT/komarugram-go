// SPDX-License-Identifier: Unlicense

package chatmedia

import (
	"context"
	"sync"
)

type closer interface{ Close(context.Context) error }

// codec is a sandboxed decoder runtime shared by the workers of a manager.
// Compiling it is expensive, so it is kept between animations, but it holds
// compiled machine code outside the Go heap: retire lets a hidden window give
// it back once the workers still using it are gone.
type codec[T closer] struct {
	mu      sync.Mutex
	open    func(context.Context) (T, error)
	current *lease[T]
}

type lease[T closer] struct {
	rt      T
	users   int
	retired bool
}

// acquire returns the runtime, compiling it if needed, and the function that
// releases it. The runtime stays open until released, even when retired.
func (c *codec[T]) acquire(ctx context.Context) (T, func(), error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.current == nil {
		rt, err := c.open(ctx)
		if err != nil {
			var zero T
			return zero, func() {}, err
		}
		c.current = &lease[T]{rt: rt}
	}
	l := c.current
	l.users++
	return l.rt, sync.OnceFunc(func() {
		c.mu.Lock()
		l.users--
		idle := l.retired && l.users == 0
		c.mu.Unlock()
		if idle {
			l.rt.Close(context.Background())
		}
	}), nil
}

// retire closes the runtime when its last user releases it; the next acquire
// compiles a new one.
func (c *codec[T]) retire() {
	c.mu.Lock()
	l := c.current
	c.current = nil
	idle := l != nil && l.users == 0
	if l != nil {
		l.retired = true
	}
	c.mu.Unlock()
	if idle {
		l.rt.Close(context.Background())
	}
}
