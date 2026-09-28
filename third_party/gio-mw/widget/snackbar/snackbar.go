// SPDX-License-Identifier: Unlicense OR MIT

package snackbar

import (
	"fmt"
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/overlay"
	"log"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
)

var (
	defaultSpacing         = unit.Dp(8)
	paddingStart           = unit.Dp(16)
	paddingEnd             = unit.Dp(16)
	paddingRightWithAction = unit.Dp(8)
	paddingRightWithIcon   = unit.Dp(0)
	paddingIcon            = unit.Dp(12)
	recommendedMaxLength   = 68
)

type Style struct {
	sTheme         *Theme
	supportingText string
	overlayItemId  int64
}

func (s *Style) AsOverlayItem() *overlay.Item {
	overlayItem := overlay.NewItem(s.Layout, block.GravityBottomCenter)
	s.overlayItemId = overlayItem.GetId()
	return overlayItem
}

func (s *Style) WithTheme(theme *Theme) *Style {
	s.sTheme = theme
	return s
}

func (s *Style) Layout(gtx layout.Context) layout.Dimensions {
	s.sTheme = BuildTheme(gtx)
	if s.supportingText == "" {
		return layout.Dimensions{}
	}
	return block.Padding{
		Bottom: unit.Dp(16),
		Start:  unit.Dp(16),
		End:    unit.Dp(16),
	}.Layout(gtx, s.widgetLayout)
}

func (s *Style) widgetLayout(gtx layout.Context) layout.Dimensions {
	macroOp := op.Record(gtx.Ops)
	if gtx.Constraints.Min.Y < gtx.Dp(s.sTheme.EnabledContainerOneLineHeight) {
		gtx.Constraints.Min.Y = gtx.Dp(s.sTheme.EnabledContainerOneLineHeight)
	}
	contentDim := s.childrenLayout(gtx)
	contentCallOp := macroOp.Stop()

	// Build the snackbar shape.
	baseBox := wdk.Box{
		Shape:    wdk.FromCornerShapesToken(gtx, s.sTheme.EnabledContainerShape),
		EndPoint: contentDim.Size,
	}

	// Draw snackbar elevation shadow.
	sElevation := wdk.Elevation{
		Level:       s.sTheme.EnabledContainerElevation,
		ShadowColor: s.sTheme.EnabledContainerShadowColor,
	}
	sElevation.Layout(gtx, baseBox)

	// Draw snackbar background and content.
	clipStack := baseBox.Outline(gtx).Push(gtx.Ops)
	paint.Fill(gtx.Ops, s.sTheme.EnabledContainerColor.AsNRGBA())
	contentCallOp.Add(gtx.Ops)
	clipStack.Pop()

	return contentDim
}

func (s *Style) childrenLayout(gtx layout.Context) layout.Dimensions {
	c := block.Container{
		Gravity: block.GravityMiddleStart,
	}
	return c.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Padding{
			Start: paddingStart,
			End:   paddingEnd,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			var widgets []block.Segment

			tLength := len(s.supportingText)
			if tLength > recommendedMaxLength {
				// TODO: Use slog.
				warningText := fmt.Errorf("snackbar text too long, %v > %v", tLength, recommendedMaxLength)
				log.Println(warningText)
			}
			textWidget := func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = 0
				presentation := wdk.LabelStyle{
					Alignment: text.Middle,
					Color:     s.sTheme.EnabledSupportingTextColor,
					MaxLines:  2,
					Typestyle: token.TypestyleLabelMedium,
				}
				return wdk.LayoutLabel(gtx, presentation, s.supportingText)
			}
			widgets = append(widgets, block.NewSegment(textWidget))

			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowClip,
				Expand:   false,
			}.Layout(gtx, widgets...)
		})
	})
}
