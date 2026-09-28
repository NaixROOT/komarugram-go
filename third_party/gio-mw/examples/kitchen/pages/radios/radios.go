// SPDX-License-Identifier: Unlicense OR MIT

package radios

import (
	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/card"
	"gio-mw/widget/radio"
	"image"

	"gioui.org/layout"
	"gioui.org/op/paint"
)

type Language string

var (
	page           *Page
	languageLabels = map[Language]string{
		"english": "English",
		"chinese": "Chinese",
		"spanish": "Spanish",
	}
)

type Page struct {
	toggleState      *button.Button
	radioSet         *radio.Radios[Language]
	selectedLanguage Language
}

func NewPage() router.PageWidget {
	languages := []Language{"english", "chinese", "spanish"}
	selectedLanguage := languages[0]
	radioSet := radio.NewRadios(languages, selectedLanguage, func(language Language) {
		page.selectLanguage(language)
	})
	page = &Page{
		toggleState:      button.Text(),
		radioSet:         radioSet,
		selectedLanguage: selectedLanguage,
	}
	return page
}

func NewPreview(c *card.Card) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		gtx = gtx.Disabled()
		return c.Layout(gtx,
			card.ContentCover(func(gtx layout.Context) layout.Dimensions {
				previewSize := image.Point{X: gtx.Constraints.Max.X, Y: 192}
				boxShape := wdk.Box{
					Shape: wdk.UniformCornerShapes(wdk.CornerShape{
						Kind: wdk.CornerKindRound,
						Size: 12,
					}),
					EndPoint: previewSize,
				}
				materialTheme := wdk.GetMaterialTheme(gtx)
				paint.FillShape(gtx.Ops, materialTheme.Scheme.SurfaceContainer.AsNRGBA(), boxShape.Outline(gtx))
				return block.Container{
					MinSize: previewSize,
					MaxSize: previewSize,
					Gravity: block.GravityMiddleCenter,
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					languages := []Language{"english", "chinese", "spanish"}
					radios := radio.NewRadios(languages, "chinese", func(language Language) {})
					radios.Update(gtx)
					return radios.Layout(gtx, radio.LeadingKind, languageLabels)
				})
			}),
			card.Content(func(gtx layout.Context) layout.Dimensions {
				txt := "Radios"
				return exp.BodyL(gtx, txt)
			}),
		)
	}
}

func (p *Page) selectLanguage(language Language) {
	p.selectedLanguage = language
}

func (p *Page) IsWide() bool {
	return false
}

func (p *Page) Update(gtx layout.Context) {
	p.radioSet.Update(gtx)
	if p.toggleState.Clicked(gtx) {
		p.radioSet.ToggleDisable()
	}
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Radios"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Used to select one option from a set of options"
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
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.radioSet.Layout(gtx, radio.LeadingKind, languageLabels)
		}),
		block.NewVerticalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			txt := "Selected: " + languageLabels[p.selectedLanguage]
			return exp.BodyL(gtx, txt)
		}),
	)
}
