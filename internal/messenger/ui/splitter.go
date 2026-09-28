// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"

	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/unit"
)

// Chat list widths, modeled on Telegram Desktop: dragging the list narrower
// than listNarrowBelow collapses it to a column of avatars.
const (
	listDefaultWidth = unit.Dp(320)
	listMinWidth     = unit.Dp(260)
	listMaxWidth     = unit.Dp(600)
	// The narrow list is as wide as an avatar with the insets of the wide
	// list on both sides.
	listNarrowWidth = 2*chatAvatarInset + chatAvatarSize
	listNarrowBelow = unit.Dp(180)
	pageMinWidth    = unit.Dp(320)
	splitterGrab    = unit.Dp(8)
)

// splitter is the draggable edge between the chat list and the page.
type splitter struct {
	drag gesture.Drag
	// handleX is where the handle was laid out in the previous frame, in
	// window coordinates; drag positions are relative to it.
	handleX int
	// grab is the distance from the grab point to the edge.
	grab int
}

// Update returns the requested edge position in window coordinates, if the
// user is dragging.
func (s *splitter) Update(gtx layout.Context, edge int) (int, bool) {
	var target int
	dragged := false
	for {
		e, ok := s.drag.Update(gtx.Metric, gtx.Source, gesture.Horizontal)
		if !ok {
			break
		}
		x := s.handleX + int(e.Position.X)
		switch e.Kind {
		case pointer.Press:
			s.grab = x - edge
		case pointer.Drag:
			target, dragged = x-s.grab, true
		}
	}
	return target, dragged
}

// Layout registers the handle centered on the edge at x (window coordinates;
// the context origin is the window origin).
func (s *splitter) Layout(gtx layout.Context, x, height int) {
	grab := gtx.Dp(splitterGrab)
	s.handleX = x - grab/2
	area := image.Rect(s.handleX, 0, s.handleX+grab, height)
	defer clip.Rect(area).Push(gtx.Ops).Pop()
	pointer.CursorColResize.Add(gtx.Ops)
	offset(gtx, area.Min, func(gtx layout.Context) layout.Dimensions {
		defer clip.Rect{Max: area.Size()}.Push(gtx.Ops).Pop()
		s.drag.Add(gtx.Ops)
		return layout.Dimensions{}
	})
}

// listWidth turns a requested width into the list width in Dp and whether
// the list is narrow, given the space left next to the sidebar. The list
// collapses to avatars when dragged narrow, and when the window is too
// narrow for both the list and the page.
func listWidth(requested, available unit.Dp) (unit.Dp, bool) {
	if requested < listNarrowBelow || available-pageMinWidth < listMinWidth {
		return listNarrowWidth, true
	}
	maxWidth := min(listMaxWidth, available-pageMinWidth)
	return max(min(requested, maxWidth), listMinWidth), false
}
