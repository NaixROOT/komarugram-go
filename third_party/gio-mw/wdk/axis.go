// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"image"
)

type Axis int

const (
	AxisHorizontal Axis = iota
	AxisVertical
)

func (a Axis) OrientPoint(point image.Point) image.Point {
	if a == AxisHorizontal {
		return point
	}
	return image.Point{X: point.Y, Y: point.X}
}

func (a Axis) OrientShape(shape CornerShapes) CornerShapes {
	if a == AxisHorizontal {
		return shape
	}
	return CornerShapes{
		TopStart:    shape.BottomStart,
		TopEnd:      shape.TopStart,
		BottomStart: shape.BottomEnd,
		BottomEnd:   shape.TopEnd,
	}
}
