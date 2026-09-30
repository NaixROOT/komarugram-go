// SPDX-License-Identifier: Unlicense OR MIT

package appwindow

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/app"
	"gioui.org/gpu/headless"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// ownFrames makes the test one of a system where blurred windows draw their
// own frame, or of one where they do not.
func ownFrames(t *testing.T, own bool) {
	t.Helper()
	was := ownFrame
	ownFrame = own
	t.Cleanup(func() { ownFrame = was })
}

// configOf is the configuration the options make of a decorated window.
func configOf(opts []app.Option) app.Config {
	cnf := app.Config{Decorated: true}
	for _, opt := range opts {
		opt(unit.Metric{}, &cnf)
	}
	return cnf
}

func TestBlurTakesTheSystemFrameWhereWindowsOwnTheirs(t *testing.T) {
	ownFrames(t, true)
	for _, c := range []struct {
		transparent, blur, decorated bool
	}{
		{false, false, true},
		{true, false, true},
		// Blur is of a transparent window: alone it changes nothing.
		{false, true, true},
		{true, true, false},
	} {
		cnf := configOf(effectOptions(c.transparent, c.blur))
		if cnf.Transparent != c.transparent || cnf.BlurBehind != c.blur || cnf.Decorated != c.decorated {
			t.Errorf("transparent %v, blur %v: asked for transparent %v, blur %v, the system's frame %v", c.transparent, c.blur, cnf.Transparent, cnf.BlurBehind, cnf.Decorated)
		}
	}
	// Elsewhere the frame is not the window's to take off.
	ownFrames(t, false)
	if cnf := configOf(effectOptions(true, true)); !cnf.Decorated || !cnf.BlurBehind {
		t.Errorf("on a system that frames blurred windows: the system's frame %v, blur %v", cnf.Decorated, cnf.BlurBehind)
	}
}

func TestOwnFrameIsShownForAWindowWithoutTheSystems(t *testing.T) {
	ownFrames(t, true)
	var f frame
	gtx := layout.Context{Metric: unit.Metric{PxPerDp: 2, PxPerSp: 2}}
	for _, c := range []struct {
		name  string
		cnf   app.Config
		shown bool
	}{
		{"the system's frame", app.Config{Decorated: true}, false},
		{"no frame", app.Config{}, true},
		{"no frame, maximized", app.Config{Mode: app.Maximized}, true},
		{"no frame, fullscreen", app.Config{Mode: app.Fullscreen}, false},
	} {
		f.configure(c.cnf)
		want := 0
		if c.shown {
			want = 64
		}
		if f.shown != c.shown || f.height(gtx) != want {
			t.Errorf("%s: shown %v, %d pixels high, want %v and %d", c.name, f.shown, f.height(gtx), c.shown, want)
		}
		if f.maximized != (c.cnf.Mode == app.Maximized) {
			t.Errorf("%s: maximized %v", c.name, f.maximized)
		}
	}
	ownFrames(t, false)
	f.configure(app.Config{})
	if f.shown {
		t.Error("a frame of the window's own on a system that frames its windows")
	}
}

// frameContext is a frame context with the theme in it.
func frameContext(ops *op.Ops, size image.Point, dark bool) layout.Context {
	gtx := layout.Context{Ops: ops, Now: time.Now(), Constraints: layout.Exact(size), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Values: map[string]any{}}
	scheme := schemes.SchemeBaselineLight()
	if dark {
		scheme = schemes.SchemeBaselineDark()
	}
	wdk.InitMaterialThemeInContext(gtx, defaults.NewTheme(gtx, scheme))
	return gtx
}

func TestOwnFrameButtonsActOnTheWindow(t *testing.T) {
	ownFrames(t, true)
	var f frame
	f.configure(app.Config{})
	frame := func() system.Action {
		return f.layout(frameContext(new(op.Ops), image.Pt(400, 300), false), "Title", nil)
	}
	if got := frame(); got != 0 {
		t.Fatalf("actions without a click: %v", got)
	}
	for _, c := range []struct {
		maximized bool
		button    system.Action
		want      system.Action
	}{
		{false, system.ActionMinimize, system.ActionMinimize},
		{false, system.ActionMaximize, system.ActionMaximize},
		{true, system.ActionUnmaximize, system.ActionUnmaximize},
		{false, system.ActionClose, system.ActionClose},
	} {
		mode := app.Windowed
		if c.maximized {
			mode = app.Maximized
		}
		f.configure(app.Config{Mode: mode})
		frame()
		f.deco.Clickable(c.button).Click()
		if got := frame(); got != c.want {
			t.Errorf("button %v, maximized %v: the window is asked for %v, want %v", c.button, c.maximized, got, c.want)
		}
	}
	// A frame that is not shown takes no clicks.
	f.configure(app.Config{Decorated: true})
	f.deco.Clickable(system.ActionClose).Click()
	if got := frame(); got != 0 {
		t.Errorf("a frame that is not shown asked the window for %v", got)
	}
	// What the pointer did to a button is forgotten with the window hidden.
	f.configure(app.Config{})
	f.deco.Clickable(system.ActionMinimize).Click()
	f.configure(app.Config{Mode: app.Minimized})
	f.configure(app.Config{})
	if got := frame(); got != 0 {
		t.Errorf("a click from before the window was minimized asked for %v", got)
	}
}

func TestCaptionButtonColors(t *testing.T) {
	on := color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}
	white := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	if fill, icon := captionButtonColors(system.ActionClose, false, false, on); fill.A != 0 || icon != on {
		t.Errorf("the close button at rest: fill %v, glyph %v", fill, icon)
	}
	for _, state := range [][2]bool{{true, false}, {false, true}} {
		if fill, icon := captionButtonColors(system.ActionClose, state[0], state[1], on); fill != closeHover || icon != white {
			t.Errorf("the close button hovered %v, pressed %v: fill %v, glyph %v", state[0], state[1], fill, icon)
		}
	}
	hover, _ := captionButtonColors(system.ActionMinimize, true, false, on)
	press, icon := captionButtonColors(system.ActionMinimize, true, true, on)
	if hover.A == 0 || press.A <= hover.A || icon != on || hover.R != on.R {
		t.Errorf("another button: hovered %v, pressed %v, glyph %v", hover, press, icon)
	}
}

// filled is content that names the colors of the frame.
type filled struct{ fill, on color.NRGBA }

func (filled) Theme(layout.Context) *token.Theme { return nil }
func (filled) Update(layout.Context)             {}
func (filled) Layout(layout.Context)             {}
func (c filled) FrameFill(layout.Context) (fill, on color.NRGBA) {
	return c.fill, c.on
}

// TestOwnFrameIsDrawn draws the frame, and checks where: the caption is of
// the content's color, the window below it is left to the content, and the
// close button under the pointer is red. FRAME_PNG_DIR saves the frames of
// both themes to look at.
func TestOwnFrameIsDrawn(t *testing.T) {
	ownFrames(t, true)
	size := image.Pt(480, 120)
	window, err := headless.NewWindow(size.X, size.Y)
	if err != nil {
		t.Skipf("no headless window: %v", err)
	}
	defer window.Release()
	draw := func(f *frame, dark bool, content Content) *image.RGBA {
		ops := new(op.Ops)
		gtx := frameContext(ops, size, dark)
		f.layout(gtx, "KomaruGram Go — Ada Lovelace", content)
		if err := window.Frame(ops); err != nil {
			t.Fatal(err)
		}
		img := image.NewRGBA(image.Rectangle{Max: size})
		if err := window.Screenshot(img); err != nil {
			t.Fatal(err)
		}
		return img
	}
	var f frame
	f.configure(app.Config{})
	fill := color.NRGBA{R: 0x20, G: 0x40, B: 0x60, A: 0xff}
	img := draw(&f, false, filled{fill: fill, on: color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}})
	if got := color.NRGBAModel.Convert(img.At(200, 4)); got != fill {
		t.Errorf("the caption is %v, want the content's %v", got, fill)
	}
	if _, _, _, a := img.At(200, 60).RGBA(); a != 0 {
		t.Errorf("the frame painted below the caption: %v", img.At(200, 60))
	}
	// The glyphs are there: something of the glyph's color in each button.
	for i := range 3 {
		found := false
		for x := size.X - (3-i)*46; x < size.X-(2-i)*46; x++ {
			for y := range 32 {
				r, _, _, _ := img.At(x, y).RGBA()
				found = found || r > 0xc000
			}
		}
		if !found {
			t.Errorf("button %d has no glyph", i)
		}
	}
	if dir := os.Getenv("FRAME_PNG_DIR"); dir != "" {
		for name, dark := range map[string]bool{"light": false, "dark": true} {
			var f frame
			f.configure(app.Config{Mode: map[bool]app.WindowMode{false: app.Windowed, true: app.Maximized}[dark]})
			out, err := os.Create(filepath.Join(dir, "frame-"+name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(out, draw(&f, dark, nil)); err != nil {
				t.Fatal(err)
			}
			out.Close()
		}
	}
}
