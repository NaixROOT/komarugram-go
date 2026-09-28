// SPDX-License-Identifier: Unlicense OR MIT

package card

import (
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
)

type widgetStyle struct {
	cCard  *Card
	cTheme *Theme
}

func (s *widgetStyle) mainLayout(gtx layout.Context, children ...*Child) layout.Dimensions {
	// Draw contents and record macro and dimensions.
	macroOp := op.Record(gtx.Ops)
	// TODO: Set SurfaceTheme for children.
	contentDim := s.childrenLayout(gtx, children...)
	contentCallOp := macroOp.Stop()

	// Build the card shape.
	strokeWidth := float32(0)
	if s.cCard.Kind == Outlined {
		strokeWidth = float32(s.cTheme.EnabledOutlineWidth)
	}
	baseBox := wdk.Box{
		Shape:       wdk.FromCornerShapesToken(gtx, s.cTheme.EnabledContainerShape),
		EndPoint:    contentDim.Size,
		StrokeWidth: strokeWidth,
	}

	// Draw the shadow, outline and background.
	s.drawBackdrop(gtx, baseBox)

	// TODO: Add drag and drop support.
	// TODO: Add click behavior & disabled state support.

	// Draw card contents.
	clipStack := baseBox.Outline(gtx).Push(gtx.Ops)
	contentCallOp.Add(gtx.Ops)
	clipStack.Pop()

	return contentDim
}

func (s *widgetStyle) drawBackdrop(gtx layout.Context, baseBox wdk.Box) {
	if s.cCard.Kind == Outlined {
		// Draw card outline.
		paint.FillShape(
			gtx.Ops,
			s.cTheme.EnabledOutlineColor.AsNRGBA(),
			baseBox.Stroke(gtx),
		)
	} else {
		// Draw card elevation shadow.
		if s.cTheme.EnabledContainerElevation != token.ElevationLevel0 {
			cElevation := wdk.Elevation{
				Level:       s.cTheme.EnabledContainerElevation,
				ShadowColor: s.cTheme.EnabledContainerShadowColor,
			}
			cElevation.Layout(gtx, baseBox)
		}

		// Draw card background.
		paint.FillShape(
			gtx.Ops,
			s.cTheme.EnabledContainerColor.AsNRGBA(),
			baseBox.Outline(gtx),
		)
	}
}

func (s *widgetStyle) childrenLayout(gtx layout.Context, children ...*Child) layout.Dimensions {
	var widgets []block.Segment

	for _, child := range children {
		widgets = append(widgets, block.NewSegment(child.Layout))
	}

	if s.cCard.Clickable.Hovered() {
		pointer.CursorPointer.Add(gtx.Ops)
	}

	return s.cCard.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx, widgets...)
	})
}
