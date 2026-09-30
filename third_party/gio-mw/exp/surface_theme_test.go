// SPDX-License-Identifier: Unlicense OR MIT

package exp_test

import (
	"testing"

	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/exp"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// frame is a context as a window makes one for a frame: its own ops and
// Values, with the material theme in them.
func frame() layout.Context {
	gtx := layout.Context{Ops: new(op.Ops), Values: map[string]any{}}
	wdk.InitMaterialThemeInContext(gtx, defaults.NewTheme(gtx, schemes.SchemeBaselineLight()))
	return gtx
}

// TestSurfacesOfTwoWindows draws in two windows at once, as their
// goroutines do: one starting its frame must not take the surface another
// is drawing on. A stack for the whole program did.
func TestSurfacesOfTwoWindows(t *testing.T) {
	a, b := frame(), frame()
	exp.Background(a)
	container := exp.PrimaryContainer(a, clip.Rect{}.Op())
	defer container.Pop()
	exp.Background(b)
	if got := exp.GetSurfaceTheme(a); got != container {
		t.Errorf("window A draws on %v after window B began its frame, want its container %v", got.Color, container.Color)
	}
	scheme := wdk.GetMaterialTheme(b).Scheme
	if got := exp.GetSurfaceTheme(b); got.Color != scheme.Background.Color {
		t.Errorf("window B draws on %v, want its background %v", got.Color, scheme.Background.Color)
	}
}

// TestSurfaceWithoutBackground reads the surface of a frame that pushed
// none, as a window that paints its own background does: it is the
// background, not a panic.
func TestSurfaceWithoutBackground(t *testing.T) {
	gtx := frame()
	scheme := wdk.GetMaterialTheme(gtx).Scheme
	if got := exp.GetSurfaceTheme(gtx); got.Color != scheme.Background.Color || got.OnColor != scheme.Background.OnColor {
		t.Errorf("got %v on %v, want the background", got.OnColor, got.Color)
	}
	surface := exp.Surface(gtx, clip.Rect{}.Op())
	if got := exp.GetSurfaceTheme(gtx); got != surface {
		t.Error("a pushed surface is not the current one")
	}
	surface.Pop()
	if got := exp.GetSurfaceTheme(gtx); got.Color != scheme.Background.Color {
		t.Error("popping the only surface did not leave the background")
	}
}
