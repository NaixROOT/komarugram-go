// SPDX-License-Identifier: Unlicense OR MIT

package divider

import (
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"image"
)

const (
	thickness = 1
)

type Style struct {
	dTheme *Theme
}

func (s *Style) Thickness(gtx layout.Context) int {
	return gtx.Dp(unit.Dp(thickness))
}

func (s *Style) Layout(gtx layout.Context) layout.Dimensions {
	s.dTheme = BuildTheme(gtx)
	rectSize := image.Point{
		X: gtx.Constraints.Max.X,
		Y: gtx.Dp(unit.Dp(thickness)),
	}
	rect := clip.Rect{Max: rectSize}
	paint.FillShape(gtx.Ops, s.dTheme.EnabledContainerColor.AsNRGBA(), rect.Op())
	return layout.Dimensions{Size: rectSize}
}
