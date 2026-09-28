// SPDX-License-Identifier: Unlicense OR MIT

package cards

import (
	"gio-mw/examples/assets"
	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/card"
	"image"

	"gioui.org/layout"
	"gioui.org/widget"
)

type Page struct {
	galaxyImage *widget.Image
}

func NewPage() router.PageWidget {
	exampleAssets := assets.GetExampleAssets()
	galaxyImage := wdk.RequireImageFromFS(exampleAssets, "galaxy2207.png")
	galaxyImage.Fit = widget.ScaleDown

	return &Page{
		galaxyImage: galaxyImage,
	}
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
				txt := "Cards"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Used to display content and actions about a single subject"
				return exp.BodyL(gtx, txt)
			}),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.cardConfig),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.cardExamples),
		)
	})
}

func (p *Page) cardConfig(gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{}
}

func (p *Page) cardExamples(gtx layout.Context) layout.Dimensions {
	return block.Line{
		Axis:     block.AxisHorizontal,
		Overflow: block.OverflowClip,
		Expand:   true,
	}.Layout(gtx,
		block.NewFlexSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisVertical,
				Overflow: block.OverflowClip,
			}.Layout(gtx,
				block.NewFlexSegment(func(gtx layout.Context) layout.Dimensions {
					cardElevated := card.Card{Kind: card.Elevated}
					return cardElevated.Layout(gtx,
						card.Content(func(gtx layout.Context) layout.Dimensions {
							return layout.Dimensions{Size: image.Point{Y: 32}}
						}),
						card.Content(func(gtx layout.Context) layout.Dimensions {
							txt := "Elevated"
							return exp.BodyL(gtx, txt)
						}),
					)
				}),
				block.NewVerticalSpacer(card.SpacingBetweenCards),
				block.NewFlexSegment(func(gtx layout.Context) layout.Dimensions {
					cardFilled := card.Card{Kind: card.Filled}
					return cardFilled.Layout(gtx,
						card.Content(func(gtx layout.Context) layout.Dimensions {
							return layout.Dimensions{Size: image.Point{Y: 32}}
						}),
						card.Content(func(gtx layout.Context) layout.Dimensions {
							txt := "Filled"
							return exp.BodyL(gtx, txt)
						}),
					)
				}),
			)
		}),
		block.NewHorizontalSpacer(card.SpacingBetweenCards),
		block.NewFlexSegment(func(gtx layout.Context) layout.Dimensions {
			cardOutlined := card.Card{Kind: card.Outlined}
			return cardOutlined.Layout(gtx,
				card.Image(p.galaxyImage),
				card.Content(func(gtx layout.Context) layout.Dimensions {
					txt := "Outlined"
					return exp.BodyL(gtx, txt)
				}),
			)
		}),
	)
}
