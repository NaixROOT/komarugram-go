// SPDX-License-Identifier: Unlicense OR MIT

package buttons

import (
	"image"
	"komarugram/internal/kitchen/services"

	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/card"

	"gioui.org/layout"
	"gioui.org/op/paint"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type buttonStyle int

const (
	textButton buttonStyle = iota
	iconButton
	iconOnlyButton
)

var buttonIcons = []wdk.IconWidget{
	nil,
	wdk.RequireIconWidget(icons.ContentAdd),
	wdk.RequireIconWidget(icons.ContentCreate),
	wdk.RequireIconWidget(icons.NavigationMenu),
	wdk.RequireIconWidget(icons.ActionFavorite),
	wdk.RequireIconWidget(icons.ActionSettings),
}

type Page struct {
	buttonElevated    *button.Button
	buttonFilled      *button.Button
	buttonFilledTonal *button.Button
	buttonOutlined    *button.Button
	buttonText        *button.Button
	disableButtons    bool
	toggleIcon        *button.Button
	toggleStyle       *button.Button
	toggleState       *button.Button
	toggleWide        *button.Button
	buttonIcon        int
	buttonStyle       buttonStyle
	wide              bool
}

func NewPage() router.PageWidget {
	p := &Page{
		buttonElevated:    button.Elevated(),
		buttonFilled:      button.Filled(),
		buttonFilledTonal: button.FilledTonal(),
		buttonOutlined:    button.Outlined(),
		buttonText:        button.Text(),
		toggleIcon:        button.Text(),
		toggleState:       button.Text(),
		toggleStyle:       button.Text(),
		toggleWide:        button.Text(),
	}
	if p.buttonStyle == textButton {
		p.toggleIcon.Disable()
	}
	return p
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
					return block.UniformPadding(examples.SpacingSmall).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return block.Line{
							Axis:     block.AxisHorizontal,
							Overflow: block.OverflowWrap,
							Expand:   true,
						}.Layout(gtx,
							block.NewSegment(func(gtx layout.Context) layout.Dimensions {
								return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									buttonElevated := button.Elevated()
									return buttonElevated.LayoutWithIcon(gtx, "Elevated", buttonIcons[1])
								})
							}),
							block.NewSegment(func(gtx layout.Context) layout.Dimensions {
								return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									buttonElevated := button.Filled()
									return buttonElevated.Layout(gtx, "Filled")
								})
							}),
							block.Segment{BaseSize: gtx.Metric.PxToDp(gtx.Constraints.Max.X), Widget: func(gtx layout.Context) layout.Dimensions {
								return layout.Dimensions{Size: image.Point{X: gtx.Constraints.Max.X}}
							}},
							block.NewSegment(func(gtx layout.Context) layout.Dimensions {
								return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									buttonElevated := button.FilledTonal()
									return buttonElevated.Layout(gtx, "Filled Tonal")
								})
							}),
							block.NewSegment(func(gtx layout.Context) layout.Dimensions {
								return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									buttonElevated := button.Outlined()
									return buttonElevated.LayoutIconOnly(gtx, "Outlined", buttonIcons[3])
								})
							}),
							block.NewSegment(func(gtx layout.Context) layout.Dimensions {
								return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
									buttonElevated := button.Text()
									return buttonElevated.LayoutIconOnly(gtx, "Text", buttonIcons[2])
								})
							}),
						)
					})
				})
			}),
			card.Content(func(gtx layout.Context) layout.Dimensions {
				txt := "Buttons"
				return exp.BodyL(gtx, txt)
			}),
		)
	}
}

func (p *Page) HasPreview() bool {
	return true
}

func (p *Page) IsWide() bool {
	return p.wide
}

func (p *Page) Update(gtx layout.Context) {
	notificationService := services.GetNotificationService()
	if p.buttonElevated.Clicked(gtx) {
		notificationService.ShowNotification("Elevated button clicked")
	}
	if p.buttonFilled.Clicked(gtx) {
		notificationService.ShowNotification("Filled button clicked")
	}
	if p.buttonFilledTonal.Clicked(gtx) {
		notificationService.ShowNotification("Filled Tonal button clicked")
	}
	if p.buttonOutlined.Clicked(gtx) {
		notificationService.ShowNotification("Outlined button clicked")
	}
	if p.buttonText.Clicked(gtx) {
		notificationService.ShowNotification("Text button clicked")
	}
	if p.toggleState.Clicked(gtx) {
		p.disableButtons = !p.disableButtons
		if p.disableButtons {
			p.buttonElevated.Disable()
			p.buttonFilled.Disable()
			p.buttonFilledTonal.Disable()
			p.buttonOutlined.Disable()
			p.buttonText.Disable()
		} else {
			p.buttonElevated.Enable()
			p.buttonFilled.Enable()
			p.buttonFilledTonal.Enable()
			p.buttonOutlined.Enable()
			p.buttonText.Enable()
		}
	}
	if p.toggleStyle.Clicked(gtx) {
		p.buttonStyle = (p.buttonStyle + 1) % 3
		if p.buttonStyle > 0 && p.buttonIcon == 0 {
			p.buttonIcon = 1
		}
		if p.buttonStyle == 0 {
			p.toggleIcon.Disable()
		} else {
			p.toggleIcon.Enable()
		}
	}
	if p.toggleIcon.Clicked(gtx) {
		p.buttonIcon = (p.buttonIcon + 1) % len(buttonIcons)
		if p.buttonStyle > 0 && p.buttonIcon == 0 {
			p.buttonIcon = 1
		}
	}
	if p.toggleWide.Clicked(gtx) {
		p.wide = !p.wide
	}
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		var widgets []block.Segment
		widgets = append(widgets,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Buttons"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Used to prompt for actions in a UI"
				return exp.BodyL(gtx, txt)
			}),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.buttonConfig),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.buttonExamples),
		)

		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx, widgets...)
	})
}

func (p *Page) buttonConfig(gtx layout.Context) layout.Dimensions {
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowWrap,
				Expand:   true,
			}.Layout(gtx,
				block.NewSegment(func(gtx layout.Context) layout.Dimensions {
					return p.toggleState.Layout(gtx, "Toggle State")
				}),
				block.NewSegment(func(gtx layout.Context) layout.Dimensions {
					return p.toggleStyle.Layout(gtx, "Toggle Style")
				}),
				block.NewSegment(func(gtx layout.Context) layout.Dimensions {
					return p.toggleIcon.Layout(gtx, "Toggle Icon")
				}),
				block.NewSegment(func(gtx layout.Context) layout.Dimensions {
					return p.toggleWide.Layout(gtx, "Toggle Wide")
				}),
			)
		}),
	)
}

func (p *Page) buttonExamples(gtx layout.Context) layout.Dimensions {
	buttonIcon := buttonIcons[p.buttonIcon]

	buttonWidgets := []layout.Widget{
		func(gtx layout.Context) layout.Dimensions {
			return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.buttonExample(gtx, p.buttonElevated, "Elevated", buttonIcon)
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.buttonExample(gtx, p.buttonFilled, "Filled", buttonIcon)
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.buttonExample(gtx, p.buttonFilledTonal, "Filled Tonal", buttonIcon)
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.buttonExample(gtx, p.buttonOutlined, "Outlined", buttonIcon)
			})
		},
		func(gtx layout.Context) layout.Dimensions {
			return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return p.buttonExample(gtx, p.buttonText, "Text", buttonIcon)
			})
		},
	}

	var buttonSegments []block.Segment
	for _, buttonWidget := range buttonWidgets {
		if p.wide {
			buttonSegments = append(buttonSegments, block.NewFlexSegment(buttonWidget))
		} else {
			buttonSegments = append(buttonSegments, block.NewSegment(buttonWidget))
		}
	}

	if p.wide {
		return block.Line{
			Axis:     block.AxisHorizontal,
			Overflow: block.OverflowClip,
		}.Layout(gtx, buttonSegments...)
	}

	return block.Line{
		Axis:     block.AxisHorizontal,
		Overflow: block.OverflowWrap,
		Expand:   true,
	}.Layout(gtx, buttonSegments...)
}

func (p *Page) buttonExample(gtx layout.Context, btn *button.Button, label string, icon wdk.IconWidget) layout.Dimensions {
	if p.buttonStyle == iconButton {
		return btn.LayoutWithIcon(gtx, label, icon)
	}
	if p.buttonStyle == iconOnlyButton {
		return btn.LayoutIconOnly(gtx, label, icon)

	}
	return btn.Layout(gtx, label)
}
