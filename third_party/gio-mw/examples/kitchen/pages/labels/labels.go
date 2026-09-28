// SPDX-License-Identifier: Unlicense OR MIT

package labels

import (
	"gio-mw/examples/kitchen/components"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/token"
	"gio-mw/wdk/block"

	"gioui.org/layout"
)

type Page struct {
}

func NewPage() router.PageWidget {
	return &Page{}
}

func (p *Page) IsWide() bool {
	return false
}

func (p *Page) Update(gtx layout.Context) {
	// Nothing to do.
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return p.getStyleListWidget(gtx)
	})
}

func (p *Page) getStyleListWidget(gtx layout.Context) layout.Dimensions {
	var widgets []block.Segment
	widgets = append(widgets,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowWrap,
				Expand:   true,
			}.Layout(gtx,
				components.NewLabelBox(token.TypestyleDisplayLarge),
				components.NewLabelBox(token.TypestyleDisplayMedium),
				components.NewLabelBox(token.TypestyleDisplaySmall),
			)
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowWrap,
				Expand:   true,
			}.Layout(gtx,
				components.NewLabelBox(token.TypestyleHeadlineLarge),
				components.NewLabelBox(token.TypestyleHeadlineMedium),
				components.NewLabelBox(token.TypestyleHeadlineSmall),
			)
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowWrap,
				Expand:   true,
			}.Layout(gtx,
				components.NewLabelBox(token.TypestyleTitleLarge),
				components.NewLabelBox(token.TypestyleTitleMedium),
				components.NewLabelBox(token.TypestyleTitleSmall),
			)
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowWrap,
				Expand:   true,
			}.Layout(gtx,
				components.NewLabelBox(token.TypestyleBodyLarge),
				components.NewLabelBox(token.TypestyleBodyMedium),
				components.NewLabelBox(token.TypestyleBodySmall),
			)
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowWrap,
				Expand:   true,
			}.Layout(gtx,
				components.NewLabelBox(token.TypestyleLabelLarge),
				components.NewLabelBox(token.TypestyleLabelMedium),
				components.NewLabelBox(token.TypestyleLabelSmall),
			)
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowWrap,
				Expand:   true,
			}.Layout(gtx,
				components.NewLabelBox(token.TypestyleDisplayLargeEmphasized),
				components.NewLabelBox(token.TypestyleDisplayMediumEmphasized),
				components.NewLabelBox(token.TypestyleDisplaySmallEmphasized),
			)
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowWrap,
				Expand:   true,
			}.Layout(gtx,
				components.NewLabelBox(token.TypestyleHeadlineLargeEmphasized),
				components.NewLabelBox(token.TypestyleHeadlineMediumEmphasized),
				components.NewLabelBox(token.TypestyleHeadlineSmallEmphasized),
			)
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowWrap,
				Expand:   true,
			}.Layout(gtx,
				components.NewLabelBox(token.TypestyleTitleLargeEmphasized),
				components.NewLabelBox(token.TypestyleTitleMediumEmphasized),
				components.NewLabelBox(token.TypestyleTitleSmallEmphasized),
			)
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowWrap,
				Expand:   true,
			}.Layout(gtx,
				components.NewLabelBox(token.TypestyleBodyLargeEmphasized),
				components.NewLabelBox(token.TypestyleBodyMediumEmphasized),
				components.NewLabelBox(token.TypestyleBodySmallEmphasized),
			)
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowWrap,
				Expand:   true,
			}.Layout(gtx,
				components.NewLabelBox(token.TypestyleLabelLargeEmphasized),
				components.NewLabelBox(token.TypestyleLabelMediumEmphasized),
				components.NewLabelBox(token.TypestyleLabelSmallEmphasized),
			)
		}),
	)
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx, widgets...)
}
