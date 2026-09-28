// SPDX-License-Identifier: Unlicense OR MIT

package shell

import (
	"context"
	"gioui.org/io/system"
	"sync"
)

type Application struct {
	Locale system.Locale

	Context  context.Context
	Shutdown context.CancelFunc
	active   sync.WaitGroup
}

func NewApplication(ctx context.Context) *Application {
	ctx, cancel := context.WithCancel(ctx)
	return &Application{
		Context:  ctx,
		Shutdown: cancel,
	}
}

// Wait waits for all windows to close.
func (a *Application) Wait() {
	a.active.Wait()
}
