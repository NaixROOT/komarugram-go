// SPDX-License-Identifier: Unlicense OR MIT

package block

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type Background struct {
	CornerShapes wdk.CornerShapes
	Color        token.MatColor
	Elevation    token.ElevationLevel
}

func (b Background) Layout(gtx layout.Context, widget layout.Widget) layout.Dimensions {
	macroOp := op.Record(gtx.Ops)
	dimensions := widget(gtx)
	callOp := macroOp.Stop()

	materialTheme := wdk.GetMaterialTheme(gtx)

	if b.CornerShapes.IsZero() {
		if b.Elevation != token.ElevationLevel0 {
			baseBox := wdk.Box{
				EndPoint: dimensions.Size,
			}
			dElevation := wdk.Elevation{
				Level:       b.Elevation,
				ShadowColor: materialTheme.Scheme.Shadow,
			}
			dElevation.Layout(gtx, baseBox)
		}
		paint.FillShape(gtx.Ops, b.Color.AsNRGBA(), clip.Rect{Max: dimensions.Size}.Op())
	} else {
		baseBox := wdk.Box{
			Shape:    b.CornerShapes,
			EndPoint: dimensions.Size,
		}
		if b.Elevation != token.ElevationLevel0 {
			dElevation := wdk.Elevation{
				Level:       b.Elevation,
				ShadowColor: materialTheme.Scheme.Shadow,
			}
			dElevation.Layout(gtx, baseBox)
		}
		paint.FillShape(gtx.Ops, b.Color.AsNRGBA(), baseBox.Outline(gtx))
	}

	callOp.Add(gtx.Ops)

	return dimensions
}
