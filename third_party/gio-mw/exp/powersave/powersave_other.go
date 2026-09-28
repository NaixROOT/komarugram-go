// SPDX-License-Identifier: Unlicense OR MIT

//go:build !(linux && !android) && !windows

package powersave

func startPlatform(m *Monitor) func() {
	return nil
}
