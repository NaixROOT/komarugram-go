// SPDX-License-Identifier: Unlicense OR MIT

// Package alert tells the user about a failure with a message box of the
// system. It is for what a window of the application cannot tell: a failure
// before the first window, or of the window itself.
package alert

import "sync"

var (
	mu    sync.Mutex
	shown = map[string]bool{}
)

// Error shows text in a message box titled title and waits until it is
// dismissed. It reports whether a box was shown: a system without one has
// only the log. A text shown before is not shown again, so that the windows
// failing alike tell it once.
func Error(title, text string) bool {
	mu.Lock()
	again := shown[text]
	shown[text] = true
	mu.Unlock()
	if again {
		return false
	}
	return show(title, text)
}
