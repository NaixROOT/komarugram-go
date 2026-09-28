// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"

	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

const (
	navButtonWidth  = unit.Dp(76)
	navButtonHeight = unit.Dp(64)
	navIconSize     = unit.Dp(28)
	navButtonRadius = unit.Dp(12)
)

// navButton is a nearly square, transparent sidebar button with a large icon and a
// label under it. The active one is highlighted.
type navButton struct {
	surface
	content wdk.ColorTween
}

func (b *navButton) Clicked(gtx layout.Context) bool {
	return b.click.Clicked(gtx)
}

// Layout draws the button; badge is the unread count shown over the icon,
// none when 0.
func (b *navButton) Layout(gtx layout.Context, icon wdk.IconWidget, txt string, active bool, badge int) layout.Dimensions {
	sc := scheme(gtx)
	size := image.Pt(gtx.Dp(navButtonWidth), gtx.Dp(navButtonHeight))
	background := sc.SecondaryContainer.Color.SetOpacity(0)
	content := sc.SurfaceVariant.OnColor
	if active {
		background = sc.SecondaryContainer.Color
		content = sc.SecondaryContainer.OnColor
	}
	content = b.content.Animate(gtx, content)
	style := surfaceStyle{radius: gtx.Dp(navButtonRadius), background: background, content: content}
	return b.surface.Layout(gtx, size, style, func(gtx layout.Context) layout.Dimensions {
		// Icon and label, centered in the square.
		iconPx := gtx.Dp(navIconSize)
		macro := op.Record(gtx.Ops)
		labelGtx := gtx
		labelGtx.Constraints = layout.Constraints{Max: image.Pt(size.X-gtx.Dp(4), size.Y)}
		labelDims := centeredLabel(labelGtx, txt, token.TypestyleLabelSmall, content, 1)
		labelCall := macro.Stop()
		gap := gtx.Dp(4)
		top := (size.Y - iconPx - gap - labelDims.Size.Y) / 2
		iconOrigin := image.Pt((size.X-iconPx)/2, top)
		offset(gtx, iconOrigin, func(gtx layout.Context) layout.Dimensions {
			return exact(gtx, image.Pt(iconPx, iconPx), func(gtx layout.Context) layout.Dimensions {
				return icon(gtx, content)
			})
		})
		offset(gtx, image.Pt((size.X-labelDims.Size.X)/2, top+iconPx+gap), func(gtx layout.Context) layout.Dimensions {
			labelCall.Add(gtx.Ops)
			return labelDims
		})
		if badge > 0 {
			drawBadge(gtx, image.Pt(iconOrigin.X+iconPx-gtx.Dp(6), iconOrigin.Y-gtx.Dp(4)), badge, sc.Primary.Color, sc.Primary.OnColor)
		}
		return layout.Dimensions{Size: size}
	})
}
