// SPDX-License-Identifier: Unlicense OR MIT

package sheet

import (
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/overlay"
	"image"

	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

type Position uint8

const (
	Bottom Position = iota
	Side
)

type Sheet struct {
	content      *layout.List
	handle       *gesture.Drag
	position     Position
	overlayState *overlayState
}

type overlayState struct {
	item         *overlay.Item
	dragDistance int
	dragStart    int
}

func (s *Sheet) update(gtx layout.Context) {
	totalDrag := 0
	for {
		dragAxis := gesture.Horizontal
		if s.position == Bottom {
			dragAxis = gesture.Vertical
		}
		update, ok := s.handle.Update(gtx.Metric, gtx.Source, dragAxis)
		if !ok {
			break
		}
		dragPosition := 0
		if s.position == Bottom {
			dragPosition = int(update.Position.Y)
		} else {
			dragPosition = int(update.Position.X)
		}
		if update.Kind == pointer.Press {
			s.overlayState.dragStart = s.overlayState.dragStart + dragPosition
		} else if update.Kind == pointer.Drag {
			totalDrag = s.overlayState.dragStart - dragPosition
		} else if update.Kind == pointer.Release {
			totalDrag = s.overlayState.dragStart - dragPosition
			s.overlayState.dragStart = 0
		}
	}
	s.overlayState.dragDistance += totalDrag
	if s.overlayState.dragDistance >= 0 {
		s.overlayState.dragDistance = 0
	} else {
		minSheetSize := gtx.Dp(50)
		dragLowerBound := -(s.getMaxSize(gtx) - minSheetSize)
		if s.overlayState.dragDistance < dragLowerBound {
			s.overlayState.dragDistance = dragLowerBound
			s.overlayState.item.Close()
			s.overlayState = nil
		}
	}
}

func (s *Sheet) getMaxSize(gtx layout.Context) int {
	if s.position == Bottom {
		return gtx.Constraints.Max.Y - gtx.Dp(72)
	}
	availableWidth := gtx.Constraints.Max.X - gtx.Dp(72)
	return min(availableWidth, gtx.Dp(maxWidth))
}

func (s *Sheet) layout(gtx layout.Context, widget layout.Widget) layout.Dimensions {
	if s.overlayState == nil {
		return layout.Dimensions{}
	}
	if s.position == Bottom {
		gtx.Constraints.Min.Y = 0
	} else {
		gtx.Constraints.Min.X = 0
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	}

	widgetTheme := BuildTheme(gtx)
	containerShape := widgetTheme.BottomDockedContainerShape
	if s.position == Side {
		containerShape = widgetTheme.SideDockedContainerShape
	}
	return block.Background{
		CornerShapes: wdk.FromCornerShapesToken(gtx, containerShape),
		Color:        widgetTheme.DockedContainerColor,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		absoluteMaxSize := s.getMaxSize(gtx) + s.overlayState.dragDistance
		if s.position == Bottom {
			gtx.Constraints.Max.Y = absoluteMaxSize
		} else {
			if gtx.Constraints.Max.X > absoluteMaxSize {
				gtx.Constraints.Max.X = absoluteMaxSize
			}
		}

		lineAxis := block.AxisHorizontal
		lineExpand := false
		if s.position == Bottom {
			lineAxis = block.AxisVertical
			lineExpand = true
		}
		return block.Line{
			Axis:     lineAxis,
			Overflow: block.OverflowClip,
			Expand:   lineExpand,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				dragHandleDimensions := s.layoutDragHandle(gtx, widgetTheme)
				defer clip.Rect{Max: dragHandleDimensions.Size}.Push(gtx.Ops).Pop()
				s.handle.Add(gtx.Ops)
				pointer.CursorGrab.Add(gtx.Ops)
				return dragHandleDimensions
			}).AlignMiddle(),
			block.NewFlexSegment(func(gtx layout.Context) layout.Dimensions {
				return s.layoutContents(gtx, widget)
			}).AlignMiddle(),
		)
	})
}

func (s *Sheet) layoutDragHandle(gtx layout.Context, widgetTheme *Theme) layout.Dimensions {
	return block.UniformPadding(unit.Dp(22)).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		handleWidth := gtx.Dp(widgetTheme.DockedDragHandleWidth)
		handleHeight := gtx.Dp(widgetTheme.DockedDragHandleHeight)
		dragHandleSize := image.Point{X: handleWidth, Y: handleHeight}
		if s.position == Side {
			dragHandleSize.X, dragHandleSize.Y = dragHandleSize.Y, dragHandleSize.X
		}
		cornerShapes := token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		})
		baseBox := wdk.Box{
			EndPoint: dragHandleSize,
			Shape:    wdk.FromCornerShapesToken(gtx, cornerShapes),
		}
		paint.FillShape(gtx.Ops, widgetTheme.DockedDragHandleColor.AsNRGBA(), baseBox.Outline(gtx))
		return layout.Dimensions{Size: dragHandleSize}
	})
}

func (s *Sheet) layoutContents(gtx layout.Context, widget layout.Widget) layout.Dimensions {
	if s.position == Bottom {
		if gtx.Constraints.Max.X > gtx.Dp(maxWidth) {
			gtx.Constraints.Min.X = gtx.Dp(maxWidth)
		} else {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
		}
	}
	return s.content.Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
		return widget(gtx)
	})
}
