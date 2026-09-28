// SPDX-License-Identifier: Unlicense OR MIT

package block

import (
	"gioui.org/io/system"
	"gioui.org/layout"
	"image"
)

var (
	english = system.Locale{
		Language:  "EN",
		Direction: system.LTR,
	}
	yidish = system.Locale{
		Language:  "YI",
		Direction: system.RTL,
	}
	constraintsDefault = layout.Constraints{
		Max: image.Point{X: 128, Y: 128},
	}
	constraintsSmall = layout.Constraints{
		Max: image.Point{X: 16, Y: 16},
	}
	constraintsZeroHeight = layout.Constraints{
		Max: image.Point{X: 128, Y: 0},
	}
	constraintsZeroWidth = layout.Constraints{
		Max: image.Point{X: 0, Y: 128},
	}
	constraintsVeryTall = layout.Constraints{
		Max: image.Point{X: 128, Y: 1000000},
	}
	constraintsVeryWide = layout.Constraints{
		Max: image.Point{X: 1000000, Y: 128},
	}
	constraintsRectWide = layout.Constraints{
		Min: image.Point{X: 32, Y: 16},
		Max: image.Point{X: 128, Y: 128},
	}
	constraintsRectTall = layout.Constraints{
		Min: image.Point{X: 16, Y: 32},
		Max: image.Point{X: 128, Y: 128},
	}
	widget016By016 = func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{
			Size:     image.Point{X: 16, Y: 16},
			Baseline: 0,
		}
	}
	widget008By000 = func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{X: 8, Y: 0}}
	}
	widget000By008 = func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{X: 0, Y: 8}}
	}
	widget008By008 = func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{X: 8, Y: 8}, Baseline: 4}
	}
	widget128By128 = func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{X: 128, Y: 128}}
	}
	widget256By256 = func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Point{X: 256, Y: 256}}
	}
)
