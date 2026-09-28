// SPDX-License-Identifier: Unlicense OR MIT

package tooltip

import (
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
)

type widgetStyle struct {
	Tooltip *Tooltip
	theme   *Theme
}

func (s *widgetStyle) layout(gtx layout.Context, targetDimensions layout.Dimensions) layout.Dimensions {
	wdk.EnforceMin(&gtx, 0, 0)

	// Draw contents and record macro and dimensions.
	macroOp := op.Record(gtx.Ops)
	buttonDim := s.drawTextLayer(gtx)
	callOp := macroOp.Stop()

	xOffset := targetDimensions.Size.X/2 - buttonDim.Size.X/2
	yOffset := -buttonDim.Size.Y - gtx.Dp(s.theme.targetSpacing)
	transformStack := op.Offset(image.Pt(xOffset, yOffset)).Push(gtx.Ops)

	// Build the tooltip shape.
	baseBox := wdk.Box{
		Shape:    wdk.FromCornerShapesToken(gtx, s.theme.EnabledContainerShape),
		EndPoint: buttonDim.Size,
	}

	// Draw the tooltip shape.
	paint.FillShape(
		gtx.Ops,
		s.theme.EnabledContainerColor.AsNRGBA(),
		baseBox.Outline(gtx),
	)

	// Draw tooltip contents.
	callOp.Add(gtx.Ops)
	transformStack.Pop()

	return buttonDim
}

func (s *widgetStyle) drawTextLayer(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.Y = gtx.Dp(s.theme.minHeight)
	maxContentWidth := gtx.Dp(s.theme.maxWidth - s.theme.paddingContent*2)
	if gtx.Constraints.Max.X < maxContentWidth {
		gtx.Constraints.Max.X = maxContentWidth
	}
	return block.Container{
		Gravity: block.GravityMiddleCenter,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Padding{
			Start: s.theme.paddingContent,
			End:   s.theme.paddingContent,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			// Draw label and record macro and dimensions.
			wdk.EnforceMin(&gtx, 0, 0)
			presentation := wdk.LabelStyle{
				Typestyle:  token.TypestyleLabelSmall,
				Color:      s.theme.EnabledSupportingTextColor,
				WrapPolicy: text.WrapWords,
			}
			return wdk.LayoutLabel(gtx, presentation, s.Tooltip.SupportingText)

		})
	})
}
