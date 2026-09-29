// SPDX-License-Identifier: Unlicense OR MIT

package appwindow

import (
	"syscall"

	"gioui.org/app"
	"gioui.org/io/event"
)

// CaptureExclusionSupported reports whether windows can be hidden from
// screen capture: Windows 10 2004 and later.
const CaptureExclusionSupported = true

var setWindowDisplayAffinity = syscall.NewLazyDLL("user32.dll").NewProc("SetWindowDisplayAffinity")

// The affinities of SetWindowDisplayAffinity.
const (
	wdaNone               = 0x00
	wdaExcludeFromCapture = 0x11
)

// viewHandle is the native window an event tells of, 0 when it is gone.
func viewHandle(e event.Event) (uintptr, bool) {
	v, ok := e.(app.Win32ViewEvent)
	return v.HWND, ok
}

func setCaptureExcluded(hwnd uintptr, on bool) error {
	affinity := uintptr(wdaNone)
	if on {
		affinity = wdaExcludeFromCapture
	}
	if ok, _, err := setWindowDisplayAffinity.Call(hwnd, affinity); ok == 0 {
		return err
	}
	return nil
}
