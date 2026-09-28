// SPDX-License-Identifier: Unlicense OR MIT

//go:build windows && (amd64 || arm64)

package tray

import "unsafe"

// The Win32 structures must match the sizes of their C declarations on
// 64-bit Windows; a mismatch makes these array lengths negative.
var (
	_ [unsafe.Sizeof(notifyIconData{}) - 976]struct{}
	_ [976 - unsafe.Sizeof(notifyIconData{})]struct{}
	_ [unsafe.Sizeof(wndClassEx{}) - 80]struct{}
	_ [80 - unsafe.Sizeof(wndClassEx{})]struct{}
	_ [unsafe.Sizeof(msg{}) - 48]struct{}
	_ [48 - unsafe.Sizeof(msg{})]struct{}
	_ [unsafe.Sizeof(bitmapInfoHeader{}) - 40]struct{}
	_ [40 - unsafe.Sizeof(bitmapInfoHeader{})]struct{}
	_ [unsafe.Sizeof(iconInfo{}) - 32]struct{}
	_ [32 - unsafe.Sizeof(iconInfo{})]struct{}
)
