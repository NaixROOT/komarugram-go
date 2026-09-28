// SPDX-License-Identifier: Unlicense OR MIT

package block

import (
	"gioui.org/layout"
	"gioui.org/op"
	"image"
)

type Stack struct {
	Gravity Gravity
}

func (s Stack) Layout(gtx layout.Context, segments ...Segment) layout.Dimensions {
	rSegments := make([]*renderedSegment, len(segments))

	stackSize := image.Point{}
	for idx, segment := range segments {
		macroOp := op.Record(gtx.Ops)
		rSegment := &renderedSegment{}
		rSegments[idx] = rSegment
		rSegment.dimensions = segment.Widget(gtx)
		rSegment.callOp = macroOp.Stop()
		if stackSize.X < rSegment.dimensions.Size.X {
			stackSize.X = rSegment.dimensions.Size.X
		}
		if stackSize.Y < rSegment.dimensions.Size.Y {
			stackSize.Y = rSegment.dimensions.Size.Y
		}
	}

	for _, rSegment := range rSegments {
		Container{
			MinSize: stackSize,
			MaxSize: stackSize,
			Gravity: s.Gravity,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			rSegment.callOp.Add(gtx.Ops)
			return rSegment.dimensions
		})
	}
	return layout.Dimensions{Size: stackSize}
}
