// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"gioui.org/layout"
	"image"
	"math"
)

const DimensionMax = math.MaxInt16

func EnforceHeight(gtx *layout.Context, yMin, yMax int) {
	if yMin < 0 {
		yMin = 0
	}
	if yMax < yMin {
		yMax = yMin
	}
	gtx.Constraints.Min.Y = yMin
	gtx.Constraints.Max.Y = yMax
}

func EnforceWidth(gtx *layout.Context, xMin, xMax int) {
	if xMin < 0 {
		xMin = 0
	}
	if xMax < xMin {
		xMax = xMin
	}
	gtx.Constraints.Min.X = xMin
	gtx.Constraints.Max.X = xMax
}

func EnforceMin(gtx *layout.Context, xMin, yMin int) {
	if xMin < 0 {
		xMin = 0
	}
	if yMin < 0 {
		yMin = 0
	}
	gtx.Constraints.Min = image.Point{
		X: xMin,
		Y: yMin,
	}
	if gtx.Constraints.Max.X < xMin {
		gtx.Constraints.Max.X = xMin
	}
	if gtx.Constraints.Max.Y < yMin {
		gtx.Constraints.Max.Y = yMin
	}
}

func EnforceMax(gtx *layout.Context, xMax, yMax int) {
	if xMax < 0 {
		xMax = 0
	}
	if yMax < 0 {
		yMax = 0
	}
	gtx.Constraints.Max = image.Point{
		X: xMax,
		Y: yMax,
	}
	if gtx.Constraints.Min.X > xMax {
		gtx.Constraints.Min.X = xMax
	}
	if gtx.Constraints.Min.Y > yMax {
		gtx.Constraints.Min.Y = yMax
	}
}
