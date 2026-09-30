// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"

	"gio-mw/token"
	"gioui.org/app"
	"gioui.org/layout"
)

const windowSurfaceOpacityKey = "komarugram/ui.windowSurfaceOpacity"

// The compositor blurs the desktop. Only the surface fill is translucent;
// applying an opacity layer to a widget would also fade its text and media.
func withWindowSurfaceOpacity(gtx layout.Context, transparency int, supported bool) {
	opacity := token.OpacityLevel(1)
	if supported {
		opacity = token.OpacityLevel(1 - float32(transparency)/100)
	}
	if gtx.Values != nil {
		gtx.Values[windowSurfaceOpacityKey] = opacity
	}
}

func fillWindowSurface(gtx layout.Context, fill token.MatColor, size image.Point) {
	if opacity, ok := gtx.Values[windowSurfaceOpacityKey].(token.OpacityLevel); ok {
		fill = fill.SetOpacity(opacity)
	}
	fillRect(gtx, fill, size)
}

// Keep transparency available for live changes to the slider, but request
// compositor blur only when enabled and some desktop can show through. Other backends
// ignore these options and report opaque windows through Translucency.
func (a *App) updateWindowEffects() {
	prefs := a.preferences.Global()
	want := prefs.WindowBlur && prefs.WindowTransparency > 0
	if a.windowEffectsSet && a.windowBlurWanted == want {
		return
	}
	a.windowEffectsSet, a.windowBlurWanted = true, want
	if a.window.Window != nil {
		a.window.Option(app.Transparent(true), app.BlurBehind(want))
	}
}
