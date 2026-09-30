// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"

	"gio-mw/token"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
)

// The overlays are the panels drawn over the messenger's content: the
// floating composer, the context menus and the toasts. An overlay may blur
// what is behind it and let some of it show through (preferences.Overlays,
// and the composer's own switch): it draws again, blurred, a recording of
// what it covers, which its owner makes while drawing that (history.go,
// chatlist.go), and then its own fill over it, translucent.

// overlayPrefs is how the overlays are to be drawn now.
type overlayPrefs struct {
	// menus and toasts say whether they blur, opacity how opaque the
	// overlays that blur are.
	menus, toasts bool
	opacity       float32
}

// blurBackdrop is a recording of what is under an overlay, and how the
// overlay draws over it.
type blurBackdrop struct {
	call op.CallOp
	// at is where the recording's origin is in the coordinates the overlay
	// is drawn in.
	at      image.Point
	opacity token.OpacityLevel
}

// newBackdrop is a recording call, whose origin is that of the coordinates
// it was made in, with the overlay opacity the preferences give.
func newBackdrop(call op.CallOp, opacity float32) *blurBackdrop {
	return &blurBackdrop{call: call, opacity: token.OpacityLevel(opacity)}
}

// shifted is b for overlays drawn in coordinates whose origin is by from
// the recording's: a menu of a page's header, say, drawn over the page while
// the recording is of its body.
func (b *blurBackdrop) shifted(by image.Point) *blurBackdrop {
	if b == nil {
		return nil
	}
	c := *b
	c.at = c.at.Add(by)
	return &c
}

// overlayFill fills the current clip of an overlay of size, whose top-left
// corner is at origin, with fill: over the blurred backdrop and translucent
// when there is one, opaque when there is none.
func overlayFill(gtx layout.Context, b *blurBackdrop, size, origin image.Point, fill token.MatColor, radius int) {
	if b != nil {
		layoutBackdrop(gtx, size, origin.Sub(b.at), b.call)
		fill = fill.SetOpacity(b.opacity)
	}
	fillRounded(gtx, fill, size, radius)
}

// overlayPlate is overlayFill in a clip of its own: the rounded plate of an
// overlay whose owner did not clip it.
func overlayPlate(gtx layout.Context, b *blurBackdrop, size, origin image.Point, fill token.MatColor, radius int) {
	defer clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops).Pop()
	overlayFill(gtx, b, size, origin, fill, radius)
}
