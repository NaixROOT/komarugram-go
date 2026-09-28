// SPDX-License-Identifier: Unlicense OR MIT

//go:build windows || darwin

package appearance

import "time"

const pollInterval = 5 * time.Second

// poll calls read now and every pollInterval until the returned function is
// called.
func poll(m *Monitor, read func() Scheme) func() {
	m.set(read())
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				m.set(read())
			}
		}
	}()
	var once bool
	return func() {
		if !once {
			once = true
			close(done)
		}
	}
}
