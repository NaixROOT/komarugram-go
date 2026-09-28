// SPDX-License-Identifier: Unlicense OR MIT

//go:build !(linux && !android) && !windows && !darwin

package appearance

func startPlatform(m *Monitor) func() {
	return nil
}
