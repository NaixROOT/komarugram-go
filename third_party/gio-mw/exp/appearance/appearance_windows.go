// SPDX-License-Identifier: Unlicense OR MIT

package appearance

import (
	"syscall"
	"unsafe"
)

var procRegGetValue = syscall.NewLazyDLL("advapi32.dll").NewProc("RegGetValueW")

const (
	hkeyCurrentUser = 0x80000001
	rrfRtRegDword   = 0x00000010
)

func startPlatform(m *Monitor) func() {
	return poll(m, readScheme)
}

func readScheme() Scheme {
	subkey, _ := syscall.UTF16PtrFromString(`Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`)
	value, _ := syscall.UTF16PtrFromString("AppsUseLightTheme")
	var data, size uint32 = 0, 4
	r, _, _ := procRegGetValue.Call(hkeyCurrentUser, uintptr(unsafe.Pointer(subkey)), uintptr(unsafe.Pointer(value)),
		rrfRtRegDword, 0, uintptr(unsafe.Pointer(&data)), uintptr(unsafe.Pointer(&size)))
	if r != 0 {
		return Unknown
	}
	if data == 0 {
		return Dark
	}
	return Light
}
