// SPDX-License-Identifier: Unlicense OR MIT

package checkboxes

import (
	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/card"
	"gio-mw/widget/checkbox"
	"image"
	"log"

	"gioui.org/layout"
	"gioui.org/op/paint"
)

type Ingredient string

var (
	page             *Page
	ingredientLabels = map[Ingredient]string{
		"pickles": "Pickles",
		"tomato":  "Tomato",
		"lettuce": "Lettuce",
	}
)

type Page struct {
	toggleState         *button.Button
	toggleError         *button.Button
	checkboxes          *checkbox.Checkboxes[Ingredient]
	selectedIngredients []Ingredient
}

func NewPage() router.PageWidget {
	ingredients := []Ingredient{
		"pickles",
		"tomato",
		"lettuce",
	}
	selectedIngredients := []Ingredient{
		"pickles",
		"lettuce",
	}
	checkboxes := checkbox.NewCheckboxes(ingredients, selectedIngredients, func(ingredients []Ingredient) {
		log.Printf("Selected ingredients: %v", ingredients)
		page.selectedIngredients = ingredients
	})
	page = &Page{
		toggleState:         button.Text(),
		toggleError:         button.Text(),
		checkboxes:          checkboxes,
		selectedIngredients: selectedIngredients,
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
					ingredients := []Ingredient{
						"pickles",
						"tomato",
						"lettuce",
					}
					selectedIngredients := []Ingredient{
						"pickles",
						"lettuce",
					}
					checkboxes := checkbox.NewCheckboxes(ingredients, selectedIngredients, func(ingredients []Ingredient) {})
					checkboxes.Update(gtx)
					return checkboxes.LayoutWithParent(gtx, "Additions", ingredientLabels)
				})
			}),
			card.Content(func(gtx layout.Context) layout.Dimensions {
				txt := "Checkboxes"
				return exp.BodyL(gtx, txt)
			}),
		)
	}
}

func (p *Page) IsWide() bool {
	return false
}

func (p *Page) Update(gtx layout.Context) {
	p.checkboxes.Update(gtx)
	if p.toggleState.Clicked(gtx) {
		p.checkboxes.ToggleDisable()
	}
	if p.toggleError.Clicked(gtx) {
		p.checkboxes.ToggleError()
	}
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Checkboxes"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Used to select one or more items from a list, or turn an item on or off"
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
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.checkboxes.LayoutWithParent(gtx, "Additions", ingredientLabels)
		}),
		block.NewVerticalSpacer(examples.SpacingSmall),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			txt := "Selected:"
			for idx, ingredient := range page.selectedIngredients {
				if idx == 0 {
					txt += " " + ingredientLabels[ingredient]
				} else {
					txt += ", " + ingredientLabels[ingredient]
				}
			}
			return exp.BodyL(gtx, txt)
		}),
	)
}
