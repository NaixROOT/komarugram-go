// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"gio-mw/token"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
)

type Elevation struct {
	Level       token.ElevationLevel
	ShadowColor token.MatColor
}

func (e Elevation) Layout(gtx layout.Context, baseBox Box) {
	DrawShadow(gtx, baseBox, e.ShadowColor, ShadowSize(e.Level))
}

// DrawShadow draws the shadow of baseBox with the given size; see ShadowSize.
// Sizes between levels allow animating the elevation.
func DrawShadow(gtx layout.Context, baseBox Box, shadowColor token.MatColor, shadowSize float32) {
	if shadowSize <= 0 {
		return
	}

	// HACK: Multiple scaled layers are used to draw a fake shadow for the given shape.
	macroOp := op.Record(gtx.Ops)
	paint.FillShape(
		gtx.Ops,
		shadowColor.SetOpacity(0.12).AsNRGBA(),
		baseBox.Outline(gtx),
	)
	callOp := macroOp.Stop()

	var stack op.TransformStack
	shadowShapeBounds := f32.Point{
		X: float32(baseBox.EndPoint.X - baseBox.StartPoint.X),
		Y: float32(baseBox.EndPoint.Y - baseBox.StartPoint.Y),
	}
	shadowLayersCount := float32(8)
	for layerIndex := shadowLayersCount; layerIndex > 0; layerIndex-- {
		sWidth := 0.75 + shadowSize*layerIndex*0.4/shadowLayersCount
		finalSize := shadowShapeBounds.Add(f32.Point{X: sWidth, Y: sWidth})
		scaleFactor := f32.Pt(finalSize.X/shadowShapeBounds.X, finalSize.Y/shadowShapeBounds.Y)
		xOffset := (shadowShapeBounds.X - finalSize.X) / 2
		yOffset := sWidth - 0.75
		scaleOrigin := f32.Point{X: scaleFactor.X / 2, Y: 0}
		sOffset := f32.Pt(xOffset, yOffset)
		stack = op.Affine(f32.AffineId().Offset(sOffset).Scale(scaleOrigin, scaleFactor)).Push(gtx.Ops)
		callOp.Add(gtx.Ops)
		stack.Pop()
	}
}

// ShadowSize returns the shadow size used for an elevation level.
func ShadowSize(level token.ElevationLevel) float32 {
	switch level {
	case token.ElevationLevel0:
		return 0
	case token.ElevationLevel1:
		return 1
	case token.ElevationLevel2:
		return 3
	case token.ElevationLevel3:
		return 5
	case token.ElevationLevel4:
		return 8
	case token.ElevationLevel5:
		return 12
	default:
		panic("invalid shadow draw layer")
	}
}
