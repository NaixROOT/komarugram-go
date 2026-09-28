// SPDX-License-Identifier: Unlicense OR MIT

package sliders

import (
	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/slider"

	"gioui.org/layout"
)

type Page struct {
	toggleState            *button.Button
	standardSlider         *slider.Slider
	standardVerticalSlider *slider.Slider
}

func NewPage() router.PageWidget {
	page := &Page{
		toggleState: button.Text(),
	}

	standardSliderOptions := make([]int, 11)
	for i := 0; i < len(standardSliderOptions); i++ {
		standardSliderOptions[i] = i
	}
	initialValue := 5
	page.standardSlider = slider.StandardSlider(standardSliderOptions, initialValue, func(value int) {
		page.standardVerticalSlider.SetValue(value)
	})
	page.standardVerticalSlider = slider.StandardSlider(standardSliderOptions, initialValue, func(value int) {
		page.standardSlider.SetValue(value)
	})
	page.standardVerticalSlider.Orientation = wdk.AxisVertical

	return page
}

func (p *Page) IsWide() bool {
	return false
}

func (p *Page) Update(gtx layout.Context) {
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Sliders"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Used to make selections from a range of values"
				return exp.BodyL(gtx, txt)
			}),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.sectionConfig),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.sectionExamples),
		)
	})
}

func (p *Page) sectionConfig(gtx layout.Context) layout.Dimensions {
	return block.Line{
		Axis:     block.AxisHorizontal,
		Overflow: block.OverflowWrap,
		Expand:   true,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.toggleState.Layout(gtx, "Toggle State")
		}),
	)
}

func (p *Page) sectionExamples(gtx layout.Context) layout.Dimensions {
	newSize := gtx.Dp(320)
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			wdk.EnforceWidth(&gtx, newSize, newSize)
			return p.standardSlider.Layout(gtx)
		}).AlignStart(),
		block.NewVerticalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			wdk.EnforceHeight(&gtx, newSize, newSize)
			wdk.EnforceWidth(&gtx, 0, gtx.Constraints.Max.X)
			return p.standardVerticalSlider.Layout(gtx)
		}).AlignStart(),
	)
}
