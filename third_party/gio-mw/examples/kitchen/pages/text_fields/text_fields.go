// SPDX-License-Identifier: Unlicense OR MIT

package text_fields

import (
	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/input"

	"gioui.org/layout"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type Page struct {
	toggleState *button.Button
	toggleError *button.Button
	textField1  *input.Input
	textField2  *input.Input
}

func NewPage() router.PageWidget {
	textField1 := input.FilledTextInput()
	textField1.LabelText = "Label text"
	textField1.SupportingText = "Supporting text"
	textField1.LeadingIcon = wdk.RequireIconWidget(icons.ActionSearch)
	textField2 := input.FilledTextArea()
	textField2.LabelText = "Multiline text area"
	textField2.Editor.SetText("Lorem ipsum dolor sit amet.\n\nSed do eiusmod tempor incididunt ut labore et dolore magna aliqua.")
	return &Page{
		toggleState: button.Text(),
		toggleError: button.Text(),
		textField1:  textField1,
		textField2:  textField2,
	}
}

func (p *Page) IsWide() bool {
	return false
}

func (p *Page) Update(gtx layout.Context) {
	if p.toggleState.Clicked(gtx) {
		p.textField1.Disabled = !p.textField1.Disabled
		p.textField2.Disabled = !p.textField2.Disabled
	}
	if p.toggleError.Clicked(gtx) {
		p.textField1.Error = !p.textField1.Error
		p.textField2.Error = !p.textField2.Error
		if p.textField1.Error {
			p.textField1.TrailingIcon = wdk.RequireIconWidget(icons.AlertWarning)
		} else {
			p.textField1.TrailingIcon = nil
		}
	}
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Text fields"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Used to let users enter text into a UI"
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
		block.NewHorizontalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.toggleError.Layout(gtx, "Toggle Error")
		}),
	)
}

func (p *Page) sectionExamples(gtx layout.Context) layout.Dimensions {
	wdk.EnforceMax(&gtx, 480, gtx.Constraints.Max.Y)
	txt := p.textField2.Editor.GetText()
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(p.textField1.Layout),
		block.NewVerticalSpacer(examples.SpacingMedium),
		block.NewSegment(p.textField2.Layout),
		block.NewVerticalSpacer(examples.SpacingMedium),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return exp.TitleM(gtx, txt)
		}),
	)
}
