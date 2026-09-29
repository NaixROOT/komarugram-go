// SPDX-License-Identifier: Unlicense OR MIT

package appwindow

// SetCaptureExcluded hides every window of the host from screen capture,
// as AyuGram's Streamer Mode does, where the platform can: see
// CaptureExclusionSupported. The windows apply it at their next frame.
func (h *Host) SetCaptureExcluded(on bool) {
	if h.captureExcluded.Swap(on) == on {
		return
	}
	h.mu.Lock()
	windows := make([]*Window, 0, len(h.windows))
	for w := range h.windows {
		windows = append(windows, w)
	}
	h.mu.Unlock()
	for _, w := range windows {
		w.Invalidate()
	}
}

// SetCaptureExcluded is Host.SetCaptureExcluded for the host of w.
func (w *Window) SetCaptureExcluded(on bool) {
	if w.host != nil {
		w.host.SetCaptureExcluded(on)
	}
}

// applyCapture makes w's native window what the host asks, once it has
// one. It runs on the window's goroutine.
func (w *Window) applyCapture() {
	if w.host == nil || w.view == 0 {
		return
	}
	on := w.host.captureExcluded.Load()
	if w.captureApplied && w.captureExcluded == on {
		return
	}
	if setCaptureExcluded(w.view, on) == nil {
		w.captureApplied, w.captureExcluded = true, on
	}
}
