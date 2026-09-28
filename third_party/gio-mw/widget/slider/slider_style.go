// SPDX-License-Identifier: Unlicense OR MIT

package slider

import (
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"image"
	"math"

	"gioui.org/gesture"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
)

type widgetStyle struct {
	Slider *Slider
	Theme  *Theme
}

func (s *widgetStyle) layout(gtx layout.Context) layout.Dimensions {
	switch s.Slider.sKind {
	case standardKind:
		return s.layoutStandard(gtx)
	case centeredKind:
		// TODO: Implement centered slider.
		return layout.Dimensions{}
	case rangeKind:
		// TODO: Implement range slider.
		return layout.Dimensions{}
	default:
		panic("Slider: unknown Slider kind")
	}
}

func (s *widgetStyle) layoutStandard(gtx layout.Context) layout.Dimensions {
	return s.layoutTracks(gtx)
}

func (s *widgetStyle) layoutTracks(gtx layout.Context) layout.Dimensions {
	trackHeight := gtx.Dp(s.Theme.SliderActiveTrackHeight)

	var activeStopIndicatorIndex int
	for i, opt := range s.Slider.sOptions {
		if opt == s.Slider.sStartValue {
			activeStopIndicatorIndex = i
			break
		}
	}

	stopIndicatorCount := len(s.Slider.sOptions)
	stopIndicatorSize := gtx.Dp(s.Theme.SliderStopIndicatorSize)
	stopIndicatorBox := wdk.Box{
		Shape:    wdk.FromCornerShapesToken(gtx, s.Theme.SliderStopIndicatorShape),
		EndPoint: image.Point{X: stopIndicatorSize, Y: stopIndicatorSize},
	}
	stopIndicatorSpacing := s.getStopIndicatorSpacing(gtx)
	leadingSpace := s.Theme.SliderActiveHandleLeadingSpace
	trailingSpace := s.Theme.SliderActiveHandleTrailingSpace

	segments := []block.Segment{
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			startSpace := gtx.Dp(s.Theme.SliderStopIndicatorStartSpace)
			activeTrackLength := startSpace - gtx.Dp(leadingSpace)
			if activeStopIndicatorIndex == 0 {
				return layout.Dimensions{Size: s.Slider.Orientation.OrientPoint(image.Point{X: activeTrackLength, Y: trackHeight})}
			}
			activeTrackColor := s.Theme.SliderActiveTrackColor.AsNRGBA()
			activeTrackLength += (activeStopIndicatorIndex) * stopIndicatorSize
			activeTrackLength += int(float32(activeStopIndicatorIndex) * stopIndicatorSpacing)
			activeTrackBox := wdk.Box{
				Shape: s.Slider.Orientation.OrientShape(wdk.FromCornerShapesToken(gtx, token.CornerShapes{
					TopStart:    s.Theme.SliderActiveTrackOuterCornerShape,
					BottomStart: s.Theme.SliderActiveTrackOuterCornerShape,
					TopEnd:      s.Theme.SliderActiveTrackInnerCornerShape,
					BottomEnd:   s.Theme.SliderActiveTrackInnerCornerShape,
				})),
				EndPoint: s.Slider.Orientation.OrientPoint(image.Point{X: activeTrackLength, Y: trackHeight}),
			}
			paint.FillShape(
				gtx.Ops,
				activeTrackColor,
				activeTrackBox.Outline(gtx),
			)

			yOffset := (trackHeight - stopIndicatorSize) / 2
			stopIndicatorColor := s.Theme.SliderStopIndicatorColorSelected.AsNRGBA()
			for i := range s.Slider.sOptions {
				if i >= activeStopIndicatorIndex {
					continue
				}
				xOffset := startSpace + int(stopIndicatorSpacing*float32(i)) + stopIndicatorSize*i
				if gtx.Locale.Direction == system.RTL && s.Slider.Orientation == wdk.AxisHorizontal {
					xOffset = activeTrackLength - stopIndicatorSize - xOffset
				}
				offsetPoint := image.Pt(xOffset, yOffset)
				transformStack := op.Offset(s.Slider.Orientation.OrientPoint(offsetPoint)).Push(gtx.Ops)
				paint.FillShape(gtx.Ops, stopIndicatorColor, stopIndicatorBox.Outline(gtx))
				transformStack.Pop()
			}
			return layout.Dimensions{Size: s.Slider.Orientation.OrientPoint(image.Point{X: activeTrackLength, Y: trackHeight})}
		}).AlignMiddle(),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: s.Slider.Orientation.OrientPoint(image.Point{X: gtx.Dp(leadingSpace), Y: 0})}
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			handleHeight := gtx.Dp(s.Theme.SliderHandleHeight)
			handleWidth := gtx.Dp(s.Theme.SliderHandleWidth)
			handleSize := image.Point{X: handleWidth, Y: handleHeight}
			handleOffset := image.Point{X: 0, Y: 0}
			if s.Slider.sStartDrag.Dragging() {
				activeHandleHeight := gtx.Dp(s.Theme.SliderActiveHandleHeight)
				activeHandleWidth := gtx.Dp(s.Theme.SliderActiveHandleWidth)
				handleSize = image.Point{X: activeHandleWidth, Y: activeHandleHeight}
				handleOffset.X = (handleWidth - activeHandleWidth) / 2
				handleOffset.Y = (handleHeight - activeHandleHeight) / 2
			}
			handleBox := wdk.Box{
				Shape:      s.Slider.Orientation.OrientShape(wdk.FromCornerShapesToken(gtx, s.Theme.SliderHandleShape)),
				StartPoint: s.Slider.Orientation.OrientPoint(handleOffset),
				EndPoint:   s.Slider.Orientation.OrientPoint(handleSize.Add(handleOffset)),
			}
			handleOutline := handleBox.Outline(gtx)
			defer handleOutline.Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, s.Theme.SliderHandleColor.AsNRGBA())
			return layout.Dimensions{
				Size: s.Slider.Orientation.OrientPoint(handleSize.Add(handleOffset).Add(handleOffset)),
			}
		}).AlignMiddle(),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: s.Slider.Orientation.OrientPoint(image.Point{X: gtx.Dp(trailingSpace), Y: 0})}
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			endSpace := gtx.Dp(s.Theme.SliderStopIndicatorEndSpace)
			if activeStopIndicatorIndex == stopIndicatorCount-1 {
				return layout.Dimensions{
					Size: s.Slider.Orientation.OrientPoint(image.Point{X: endSpace - gtx.Dp(trailingSpace), Y: trackHeight}),
				}
			}
			inactiveTrackLength := s.Slider.Orientation.OrientPoint(gtx.Constraints.Max).X
			activeTrackColor := s.Theme.SliderInactiveTrackColor.AsNRGBA()
			activeTrackBox := wdk.Box{
				Shape: s.Slider.Orientation.OrientShape(wdk.FromCornerShapesToken(gtx, token.CornerShapes{
					TopStart:    s.Theme.SliderActiveTrackInnerCornerShape,
					BottomStart: s.Theme.SliderActiveTrackInnerCornerShape,
					TopEnd:      s.Theme.SliderActiveTrackOuterCornerShape,
					BottomEnd:   s.Theme.SliderActiveTrackOuterCornerShape,
				})),
				EndPoint: s.Slider.Orientation.OrientPoint(image.Point{X: inactiveTrackLength, Y: trackHeight}),
			}
			paint.FillShape(
				gtx.Ops,
				activeTrackColor,
				activeTrackBox.Outline(gtx),
			)
			yOffset := (trackHeight - stopIndicatorSize) / 2
			stopIndicatorColor := s.Theme.SliderStopIndicatorColor.AsNRGBA()
			xEndOffset := inactiveTrackLength - endSpace - stopIndicatorSize
			for i := range s.Slider.sOptions {
				if i <= activeStopIndicatorIndex {
					continue
				}
				inverseIndex := stopIndicatorCount - 1 - i
				xOffset := xEndOffset - stopIndicatorSize*inverseIndex - int(stopIndicatorSpacing*float32(inverseIndex))
				if gtx.Locale.Direction == system.RTL && s.Slider.Orientation == wdk.AxisHorizontal {
					xOffset = inactiveTrackLength - endSpace - xOffset
				}
				offsetPoint := image.Pt(xOffset, yOffset)
				transformStack := op.Offset(s.Slider.Orientation.OrientPoint(offsetPoint)).Push(gtx.Ops)
				paint.FillShape(gtx.Ops, stopIndicatorColor, stopIndicatorBox.Outline(gtx))
				transformStack.Pop()
			}
			return layout.Dimensions{Size: s.Slider.Orientation.OrientPoint(image.Point{X: inactiveTrackLength, Y: trackHeight})}
		}).AlignMiddle(),
	}

	draggableValue := float32(activeStopIndicatorIndex) / float32(len(s.Slider.sOptions)-1)
	if s.Slider.Orientation == wdk.AxisVertical {
		draggableValue = 1 - draggableValue
	}
	s.Slider.sStartDrag.Value = draggableValue
	dragAxis := gesture.Horizontal
	if s.Slider.Orientation == wdk.AxisVertical {
		dragAxis = gesture.Vertical
	}
	return s.Slider.sStartDrag.Layout(gtx, dragAxis, func(gtx layout.Context) layout.Dimensions {
		if s.Slider.sStartDrag.Value != draggableValue {
			actualValue := s.Slider.sStartDrag.Value
			if s.Slider.Orientation == wdk.AxisVertical {
				actualValue = 1 - actualValue
			}
			newValueIndex := math.Round(float64(actualValue) * float64(stopIndicatorCount-1))
			s.Slider.sStartValue = s.Slider.sOptions[int(newValueIndex)]
			s.Slider.sOnChange(s.Slider.sStartValue)
		}
		lineAxis := block.AxisHorizontal
		if s.Slider.Orientation == wdk.AxisVertical {
			lineAxis = block.AxisVertical
		}
		return block.Line{
			Axis:     lineAxis,
			Overflow: block.OverflowClip,
			Expand:   true,
		}.Layout(gtx, segments...)
	})
}

func (s *widgetStyle) getStopIndicatorSpacing(gtx layout.Context) float32 {
	startSpace := gtx.Dp(s.Theme.SliderStopIndicatorStartSpace)
	endSpace := gtx.Dp(s.Theme.SliderStopIndicatorEndSpace)
	stopIndicatorCount := len(s.Slider.sOptions)
	stopIndicatorsSize := gtx.Dp(s.Theme.SliderStopIndicatorSize) * stopIndicatorCount
	availableSize := s.Slider.Orientation.OrientPoint(gtx.Constraints.Max)
	availableSpace := float32(availableSize.X - startSpace - endSpace - stopIndicatorsSize)
	return availableSpace / float32(stopIndicatorCount-1)
}
