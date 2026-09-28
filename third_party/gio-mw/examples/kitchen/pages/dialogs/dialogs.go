// SPDX-License-Identifier: Unlicense OR MIT

package dialogs

import (
	"gio-mw/examples/kitchen/services"
	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/card"
	"gio-mw/widget/dialog"
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var (
	actionDeleteForever = wdk.RequireIconWidget(icons.ActionDeleteForever)
)

type Page struct {
	showAcknowledgeDialog *button.Button
	showConfirmDialog     *button.Button
	acknowledgeDialog     *dialog.BasicStyle
	confirmDialog         *dialog.BasicStyle
}

func NewPage() router.PageWidget {
	return &Page{
		showAcknowledgeDialog: button.Text(),
		showConfirmDialog:     button.Text(),
	}
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
					return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							widget := func(gtx layout.Context) layout.Dimensions {
								gtx.Constraints.Max.X = 512
								gtx.Constraints.Max.Y = 256
								dialogInstance := dialog.Confirm(gtx).WithIcon(actionDeleteForever)
								dialogInstance.Headline = "Permanently Delete?"
								dialogInstance.Label = "Deleting the selected messages will also remove them from synced devices."
								return dialogInstance.Layout(gtx)
							}
							scaleOrigin := f32.Point{X: 0, Y: 0}
							scaleFactor := f32.Point{X: 1, Y: 1}.Div(2)
							stack := op.Affine(f32.AffineId().Scale(scaleOrigin, scaleFactor)).Push(gtx.Ops)
							dimensions := widget(gtx)
							stack.Pop()
							return layout.Dimensions{
								Size: dimensions.Size.Div(2),
							}
						})
					})
				})
			}),
			card.Content(func(gtx layout.Context) layout.Dimensions {
				txt := "Dialogs"
				return exp.BodyL(gtx, txt)
			}),
		)
	}
}

func (p *Page) IsWide() bool {
	return false
}

func (p *Page) Update(gtx layout.Context) {
	dialogService := services.GetDialogService()
	notificationService := services.GetNotificationService()
	if p.showAcknowledgeDialog.Clicked(gtx) {
		p.acknowledgeDialog = p.getAcknowledgeDialog(gtx)
		dialogService.ShowBasicDialog(p.acknowledgeDialog)
	}
	if p.showConfirmDialog.Clicked(gtx) {
		p.confirmDialog = p.getConfirmDialog(gtx)
		dialogService.ShowBasicDialog(p.confirmDialog)
	}
	if p.acknowledgeDialog != nil && p.acknowledgeDialog.ConfirmButton.Clicked(gtx) {
		notificationService.ShowNotification("OK clicked!")
		dialogService.HideBasicDialog(p.acknowledgeDialog)
		p.acknowledgeDialog = nil
	}
	if p.confirmDialog != nil && p.confirmDialog.CancelButton.Clicked(gtx) {
		notificationService.ShowNotification("Cancel clicked!")
		dialogService.HideBasicDialog(p.confirmDialog)
		p.confirmDialog = nil
	}
	if p.confirmDialog != nil && p.confirmDialog.ConfirmButton.Clicked(gtx) {
		notificationService.ShowNotification("Confirm clicked!")
		dialogService.HideBasicDialog(p.confirmDialog)
		p.confirmDialog = nil
	}
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		var widgets []block.Segment
		widgets = append(widgets,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Dialogs"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Used to provide important prompts in a user flow"
				return exp.BodyL(gtx, txt)
			}),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.dialogConfig),
		)
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx, widgets...)
	})
}

func (p *Page) dialogConfig(gtx layout.Context) layout.Dimensions {
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
					return p.showAcknowledgeDialog.Layout(gtx, "Show Acknowledge Dialog")
				}),
				block.NewHorizontalSpacer(examples.SpacingSmall),
				block.NewSegment(func(gtx layout.Context) layout.Dimensions {
					return p.showConfirmDialog.Layout(gtx, "Show Confirm Dialog")
				}),
			)
		}),
	)
}

func (p *Page) getAcknowledgeDialog(gtx layout.Context) *dialog.BasicStyle {
	dialogInstance := dialog.Acknowledge(gtx)
	dialogInstance.Headline = "Out of stock"
	dialogInstance.Label = "The item in your cart is no longer available."
	return dialogInstance
}

func (p *Page) getConfirmDialog(gtx layout.Context) *dialog.BasicStyle {
	dialogInstance := dialog.Confirm(gtx).WithIcon(actionDeleteForever)
	dialogInstance.Headline = "Permanently Delete?"
	dialogInstance.Label = "Deleting the selected messages will also remove them from synced devices."
	return dialogInstance
}
