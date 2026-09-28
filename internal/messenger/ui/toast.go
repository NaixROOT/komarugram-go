// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"time"

	"gio-mw/token"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
)

// How long a toast shows, and how long it takes to fade in and out.
const (
	toastDuration = 4 * time.Second
	toastFade     = 180 * time.Millisecond
)

// A toast is a grey plate with light text, as Telegram Desktop's: dark
// grey over a light theme, and a lighter grey over a dark one, where the
// dark grey would merge with the bubbles.
var (
	toastColor     = token.NewMatColorFromHexRGBA(0x3c3c40eb)
	toastDarkColor = token.NewMatColorFromHexRGBA(0x5a5a60f2)
	toastTextColor = token.NewMatColorFromHexRGB(0xf2f2f4)
)

// toast is a short notice, an error or the outcome of an action, on a grey
// plate at the bottom of the list or dialog it is about. It goes by itself
// after toastDuration; a new one replaces it. Every list that can have
// something to say owns one and draws it last, over its content.
type toast struct {
	text string
	// start is when it was first drawn; zero until then.
	start time.Time
}

// Show shows text, replacing what the toast showed.
func (t *toast) Show(text string) {
	t.text, t.start = text, time.Time{}
}

// Hide takes the toast away at once.
func (t *toast) Hide() { t.text = "" }

// Text is what the toast shows, "" when nothing.
func (t *toast) Text() string { return t.text }

// Layout draws the toast at the bottom of area, in the middle, and takes no
// input: what is under it stays usable.
func (t *toast) Layout(gtx layout.Context, area image.Rectangle) {
	t.layout(gtx, area, image.Rectangle{})
}

// LayoutBelow draws the toast at the top of below, in the middle, when it
// fits there: under a dialog, whose buttons it would hide at the dialog's
// bottom. Otherwise it draws at the bottom of area, the dialog.
func (t *toast) LayoutBelow(gtx layout.Context, area, below image.Rectangle) {
	t.layout(gtx, area, below)
}

func (t *toast) layout(gtx layout.Context, area, below image.Rectangle) {
	if t.text == "" || area.Empty() {
		return
	}
	if t.start.IsZero() {
		t.start = gtx.Now
	}
	age := gtx.Now.Sub(t.start)
	if age >= toastDuration {
		t.text = ""
		return
	}
	opacity := float32(1)
	switch {
	case age < toastFade:
		opacity = float32(age) / float32(toastFade)
		gtx.Execute(op.InvalidateCmd{})
	case age > toastDuration-toastFade:
		opacity = float32(toastDuration-age) / float32(toastFade)
		gtx.Execute(op.InvalidateCmd{})
	default:
		gtx.Execute(op.InvalidateCmd{At: t.start.Add(toastDuration - toastFade)})
	}

	margin := gtx.Dp(12)
	gtx.Constraints = layout.Constraints{Max: image.Pt(max(0, min(area.Dx()-2*margin, gtx.Dp(420))), max(0, area.Dy()-2*margin))}
	macro := op.Record(gtx.Ops)
	dims := layout.Inset{Top: 8, Bottom: 8, Left: 16, Right: 16}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return centeredLabel(gtx, t.text, token.TypestyleBodyMedium, toastTextColor, 4)
	})
	call := macro.Stop()
	at := image.Pt(area.Min.X+(area.Dx()-dims.Size.X)/2, area.Max.Y-margin-dims.Size.Y)
	if dims.Size.Y+margin <= below.Dy() {
		at.Y = below.Min.Y + margin
	}
	defer op.Offset(at).Push(gtx.Ops).Pop()
	defer paint.PushOpacity(gtx.Ops, opacity).Pop()
	plate := toastColor
	if s := scheme(gtx).Surface.Color; int(s.R)+int(s.G)+int(s.B) < 3*128 {
		plate = toastDarkColor
	}
	fillRounded(gtx, plate, dims.Size, gtx.Dp(12))
	call.Add(gtx.Ops)
}
