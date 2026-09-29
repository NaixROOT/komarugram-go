// SPDX-License-Identifier: Unlicense OR MIT

//go:build !windows && !linux

package deviceinfo

func model() string  { return "" }
func system() string { return "" }
