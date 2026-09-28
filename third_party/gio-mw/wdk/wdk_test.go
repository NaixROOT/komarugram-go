// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"gioui.org/layout"
	"image"
)

var (
	widget008By008 = func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{X: 8, Y: 8}}
	}
	widget032By032 = func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{X: 32, Y: 32}}
	}
	widget064By064 = func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{X: 64, Y: 64}}
	}
	widget064ByMax = func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{X: 64, Y: gtx.Constraints.Max.Y}}
	}
	widgetMaxBy064 = func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{X: gtx.Constraints.Max.X, Y: 64}}
	}
)
