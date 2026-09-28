// SPDX-License-Identifier: Unlicense OR MIT

package toggle

import (
	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/toggle"

	"gioui.org/layout"
)

type Fruit string

var (
	page        *Page
	fruitLabels = map[Fruit]string{
		"grapes":       "Grapes",
		"oranges":      "Oranges",
		"strawberries": "Strawberries",
	}
)

type Page struct {
	toggleState    *button.Button
	toggleSet      *toggle.Toggle[Fruit]
	selectedFruits []Fruit
}

func NewPage() router.PageWidget {
	fruits := []Fruit{"grapes", "oranges", "strawberries"}
	toggleSet := toggle.NewToggle(fruits, []Fruit{}, func(fruits []Fruit) {
		page.selectFruits(fruits)
	})
	page = &Page{
		toggleState: button.Text(),
		toggleSet:   toggleSet,
	}
	return page
}

func (p *Page) selectFruits(fruits []Fruit) {
	p.selectedFruits = fruits
}

func (p *Page) IsWide() bool {
	return false
}

func (p *Page) Update(gtx layout.Context) {
	if p.toggleState.Clicked(gtx) {
		p.toggleSet.ToggleDisable()
	}
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Toggles (Switches)"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Used to toggle the selection of an item on or off"
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
			return p.toggleSet.Layout(gtx, fruitLabels)
		}),
		block.NewVerticalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			txt := "Selected: "
			for idx, fruit := range p.selectedFruits {
				if idx == 0 {
					txt += fruitLabels[fruit]
				} else {
					txt += ", " + fruitLabels[fruit]
				}
			}
			return exp.BodyL(gtx, txt)
		}),
	)
}
