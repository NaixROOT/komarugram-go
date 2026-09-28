// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import "gioui.org/layout"

const AnimationsNamespace = "gio-mw.wdk.Animations"

// SetAnimationsEnabled turns widget animations on or off for the frame.
// When they are off, animated values jump straight to their targets.
// Animations are enabled unless this is called with false.
func SetAnimationsEnabled(gtx layout.Context, enabled bool) {
	gtx.Values[AnimationsNamespace] = enabled
}

// AnimationsEnabled reports whether widgets should animate in this frame.
func AnimationsEnabled(gtx layout.Context) bool {
	if enabled, ok := gtx.Values[AnimationsNamespace].(bool); ok {
		return enabled
	}
	return true
}
