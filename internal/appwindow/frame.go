// SPDX-License-Identifier: Unlicense OR MIT

package appwindow

import (
	"image"
	"image/color"
	"runtime"

	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

// On Windows the blur behind a window is acrylic, which the system draws
// behind the content only of a window without its frame, and around the
// one with it. A window that asks for blur there has no system frame and
// draws its own: a caption with the title and the window's buttons, above
// the content. Elsewhere the frame is the compositor's or Gio's.

// ownFrame is whether windows on this system draw their own frame when they
// blur what is behind them.
var ownFrame = runtime.GOOS == "windows"

func ownsFrame() bool { return ownFrame }

// effectOptions are the options of a window that is transparent, and blurs
// what is behind it.
func effectOptions(transparent, blur bool) []app.Option {
	opts := []app.Option{app.Transparent(transparent), app.BlurBehind(blur)}
	if ownsFrame() {
		opts = append(opts, app.Decorated(!(transparent && blur)))
	}
	return opts
}

// SetEffects asks for a window the desktop shows through where the content
// does not paint, and for the blur of what shows. Translucency reports what
// was granted.
func (w *Window) SetEffects(transparent, blur bool) {
	w.blurAsked = transparent && blur
	if w.Window != nil {
		w.Option(effectOptions(transparent, blur)...)
	}
}

// FrameFiller is implemented by the content whose window's own frame is not
// of the theme's surface: a translucent one, as the content is.
type FrameFiller interface {
	// FrameFill returns the color of the caption and of what is drawn on it.
	FrameFill(gtx layout.Context) (fill, on color.NRGBA)
}

const (
	captionHeight = unit.Dp(32)
	captionButton = unit.Dp(46)
	captionIcon   = unit.Dp(10)
)

// closeHover is the fill of the close button under the pointer, as the
// system's is.
var closeHover = color.NRGBA{R: 0xc4, G: 0x2b, B: 0x1c, A: 0xff}

// frame is the window's own frame.
type frame struct {
	// shown is whether the window has no frame of the system's and is not
	// fullscreen; maximized is whether it fills the screen.
	shown, maximized bool
	deco             widget.Decorations
}

// configure takes what the window is from its configuration.
func (f *frame) configure(cnf app.Config) {
	f.shown = ownsFrame() && !cnf.Decorated && cnf.Mode != app.Fullscreen
	f.maximized = cnf.Mode == app.Maximized
	if cnf.Mode == app.Minimized {
		// The pointer left with the window, and nothing told the button
		// under it.
		f.deco = widget.Decorations{}
	}
	f.deco.Maximized = f.maximized
}

// height is what the frame takes off the top of the window.
func (f *frame) height(gtx layout.Context) int {
	if !f.shown {
		return 0
	}
	return gtx.Dp(captionHeight)
}

// layout handles and draws the caption at the top of the window, and
// returns what its buttons ask of the window.
func (f *frame) layout(gtx layout.Context, title string, content Content) system.Action {
	if !f.shown {
		return 0
	}
	actions := f.deco.Update(gtx)
	sc := wdk.GetMaterialTheme(gtx).Scheme
	fill, on := sc.SurfaceContainer.AsNRGBA(), sc.Surface.OnColor.AsNRGBA()
	if filler, ok := content.(FrameFiller); ok {
		fill, on = filler.FrameFill(gtx)
	}
	size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(captionHeight))
	button := image.Pt(gtx.Dp(captionButton), size.Y)
	paint.FillShape(gtx.Ops, fill, clip.Rect{Max: size}.Op())

	// The caption moves the window, and the system maximizes it on a double
	// click there. The buttons are beside it and not over it: where the
	// window is moved, the system takes the pointer.
	bar := layout.Context(gtx)
	bar.Constraints = layout.Exact(image.Pt(max(size.X-3*button.X, 0), size.Y))
	f.deco.LayoutMove(bar, func(gtx layout.Context) layout.Dimensions {
		inset := layout.Inset{Left: unit.Dp(12), Right: unit.Dp(12)}
		return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				return wdk.LayoutLabel(gtx, wdk.LabelStyle{Typestyle: token.TypestyleLabelMedium, Color: token.MatColor(on), MaxLines: 1}, title)
			})
		})
	})

	maximize := system.ActionMaximize
	if f.maximized {
		maximize = system.ActionUnmaximize
	}
	for i, action := range []system.Action{system.ActionMinimize, maximize, system.ActionClose} {
		at := op.Offset(image.Pt(size.X-(3-i)*button.X, 0)).Push(gtx.Ops)
		click := f.deco.Clickable(action)
		gtx := gtx
		gtx.Constraints = layout.Exact(button)
		click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			back, icon := captionButtonColors(action, click.Hovered(), click.Pressed(), on)
			paint.FillShape(gtx.Ops, back, clip.Rect{Max: button}.Op())
			drawCaptionIcon(gtx, action, button, icon)
			return layout.Dimensions{Size: button}
		})
		at.Pop()
	}
	return actions
}

// captionButtonColors are the fill of a window button and the color of its
// glyph: the close button is red under the pointer, as the system's is, and
// the others are tinted with the color of the glyph.
func captionButtonColors(action system.Action, hovered, pressed bool, on color.NRGBA) (fill, icon color.NRGBA) {
	switch {
	case action == system.ActionClose && (hovered || pressed):
		return closeHover, color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case pressed:
		return withAlpha(on, 0x33), on
	case hovered:
		return withAlpha(on, 0x1f), on
	}
	return color.NRGBA{}, on
}

func withAlpha(c color.NRGBA, alpha uint8) color.NRGBA {
	c.A = alpha
	return c
}

// drawCaptionIcon draws the glyph of a window button in the middle of it:
// a line, a square, two squares or a cross, as the system's are.
func drawCaptionIcon(gtx layout.Context, action system.Action, button image.Point, c color.NRGBA) {
	side := float32(gtx.Dp(captionIcon))
	width := float32(gtx.Dp(1))
	// Whole pixels keep the lines of one pixel sharp.
	origin := f32.Pt(float32(int((float32(button.X)-side)/2)), float32(int((float32(button.Y)-side)/2)))
	stroke := func(closed bool, points ...f32.Point) {
		var path clip.Path
		path.Begin(gtx.Ops)
		path.MoveTo(points[0].Add(origin))
		for _, p := range points[1:] {
			path.LineTo(p.Add(origin))
		}
		if closed {
			path.Close()
		}
		paint.FillShape(gtx.Ops, c, clip.Stroke{Path: path.End(), Width: width}.Op())
	}
	half := width / 2
	switch action {
	case system.ActionMinimize:
		y := float32(int(side/2)) + half
		stroke(false, f32.Pt(0, y), f32.Pt(side, y))
	case system.ActionMaximize:
		stroke(true, f32.Pt(half, half), f32.Pt(side-half, half), f32.Pt(side-half, side-half), f32.Pt(half, side-half))
	case system.ActionUnmaximize:
		// The window in front, and what shows of the one behind it.
		d := float32(gtx.Dp(2))
		stroke(true, f32.Pt(half, d+half), f32.Pt(side-d-half, d+half), f32.Pt(side-d-half, side-half), f32.Pt(half, side-half))
		stroke(false, f32.Pt(d+half, d), f32.Pt(d+half, half), f32.Pt(side-half, half), f32.Pt(side-half, side-d-half), f32.Pt(side-d, side-d-half))
	case system.ActionClose:
		stroke(false, f32.Pt(0, 0), f32.Pt(side, side))
		stroke(false, f32.Pt(side, 0), f32.Pt(0, side))
	}
}
