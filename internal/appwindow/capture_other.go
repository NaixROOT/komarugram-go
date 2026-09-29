// SPDX-License-Identifier: Unlicense OR MIT

//go:build !windows

package appwindow

import (
	"errors"

	"gioui.org/io/event"
)

// CaptureExclusionSupported reports whether windows can be hidden from
// screen capture. X11 and Wayland have no way to ask for it, and macOS is
// not done yet.
const CaptureExclusionSupported = false

func viewHandle(event.Event) (uintptr, bool) { return 0, false }

func setCaptureExcluded(uintptr, bool) error { return errors.ErrUnsupported }
