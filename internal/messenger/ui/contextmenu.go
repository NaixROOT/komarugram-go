// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"

	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// menuCorner is the corner a context menu opens from: the one nearest to
// what opened it.
type menuCorner int

const (
	menuFromTopLeft menuCorner = iota
	menuFromTopRight
	menuFromBottomLeft
	menuFromBottomRight
)

// Context menus open as tdesktop's do: they fade in while they expand from
// their corner, and close the same way back, faster.
const (
	menuEnterDuration = token.DurationMedium1
	menuExitDuration  = token.DurationShort3
	// menuStartWidth and menuStartHeight are the part of the menu shown
	// when it starts to open.
	menuStartWidth  = 0.3
	menuStartHeight = 0.1
)

// contextMenu animates a panel that opens over the content from a corner:
// the emoji and sticker picker, the attachment menu and the like.
type contextMenu struct {
	visibility wdk.FloatTween
}

// Layout draws content in rect while the menu is open or animating out,
// reporting whether it drew. radius is the corner radius of the menu, which
// its expanding outline keeps.
func (m *contextMenu) Layout(gtx layout.Context, open bool, rect image.Rectangle, from menuCorner, radius int, content layout.Widget) bool {
	visibility := m.animate(gtx, open)
	if visibility == 0 {
		return false
	}
	if visibility == 1 {
		inRect(gtx, rect, content)
		return true
	}
	if !open {
		// A closing menu takes no input.
		gtx = gtx.Disabled()
	}
	size := rect.Size()
	macro := op.Record(gtx.Ops)
	inRect(gtx, image.Rectangle{Max: size}, content)
	call := macro.Stop()

	shown := image.Pt(
		int(float32(size.X)*(menuStartWidth+(1-menuStartWidth)*visibility)+.5),
		int(float32(size.Y)*(menuStartHeight+(1-menuStartHeight)*visibility)+.5),
	)
	area := image.Rectangle{Max: shown}
	if from == menuFromTopRight || from == menuFromBottomRight {
		area = area.Add(image.Pt(size.X-shown.X, 0))
	}
	if from == menuFromBottomLeft || from == menuFromBottomRight {
		area = area.Add(image.Pt(0, size.Y-shown.Y))
	}
	defer op.Offset(rect.Min).Push(gtx.Ops).Pop()
	defer clip.UniformRRect(area, min(radius, shown.X/2, shown.Y/2)).Push(gtx.Ops).Pop()
	defer paint.PushOpacity(gtx.Ops, visibility).Pop()
	call.Add(gtx.Ops)
	return true
}

// animate returns how much the menu shows, from 0 to 1.
func (m *contextMenu) animate(gtx layout.Context, open bool) float32 {
	if m.visibility.Duration == 0 {
		// Start closed, so that the menu animates in when it first opens.
		m.visibility.Animate(gtx, 0)
	}
	target := float32(0)
	m.visibility.Duration = menuExitDuration
	m.visibility.Easing = &token.EasingEmphasizedAccelerate
	if open {
		target = 1
		m.visibility.Duration = menuEnterDuration
		m.visibility.Easing = &token.EasingEmphasizedDecelerate
	}
	return m.visibility.Animate(gtx, target)
}
