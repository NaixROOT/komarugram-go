// SPDX-License-Identifier: Unlicense OR MIT

package alert

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var messageBox = windows.NewLazySystemDLL("user32.dll").NewProc("MessageBoxW")

const (
	mbIconError     = 0x00000010
	mbSetForeground = 0x00010000
)

func show(title, text string) bool {
	t, err := windows.UTF16PtrFromString(title)
	if err != nil {
		return false
	}
	m, err := windows.UTF16PtrFromString(text)
	if err != nil {
		return false
	}
	ret, _, _ := messageBox.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), mbIconError|mbSetForeground)
	return ret != 0
}
