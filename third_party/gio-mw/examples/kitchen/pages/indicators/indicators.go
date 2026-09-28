// SPDX-License-Identifier: Unlicense OR MIT

package indicators

import (
	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/indicator"

	"gioui.org/layout"
)

type Page struct {
	incrementProgress *button.Button
	resetProgress     *button.Button
	circularProgress  *indicator.Indicator
	linearProgress    *indicator.Indicator
}

func NewPage() router.PageWidget {
	return &Page{
		incrementProgress: button.Text(),
		resetProgress:     button.Text(),
		circularProgress:  indicator.Circular(),
		linearProgress:    indicator.Linear(),
	}
}

func (p *Page) IsWide() bool {
	return false
}

func (p *Page) Update(gtx layout.Context) {
	if p.incrementProgress.Clicked(gtx) {
		p.circularProgress.Progress += 0.1
		p.linearProgress.Progress += 0.1
	}
	if p.resetProgress.Clicked(gtx) {
		p.circularProgress.Progress = 0
		p.linearProgress.Progress = 0
	}
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Indicators"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Used to show the status of a process in real time"
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
			return p.incrementProgress.Layout(gtx, "Increment Progress")
		}),
		block.NewHorizontalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.resetProgress.Layout(gtx, "Reset Progress")
		}),
	)
}

func (p *Page) sectionExamples(gtx layout.Context) layout.Dimensions {
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.circularProgress.Layout(gtx)
		}),
		block.NewVerticalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.linearProgress.Layout(gtx)
		}),
	)
}
