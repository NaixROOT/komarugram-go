// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/dialog"
	"gio-mw/widget/overlay"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

func main() {
	go func() {
		var appWindow app.Window
		appWindow.Option(
			app.Title("Dialog Example"),
			app.Size(unit.Dp(420), unit.Dp(640)),
		)
		if err := appRun(&appWindow); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()
}

var (
	appOverlay            overlay.Overlay
	acknowledgeDialog     *dialog.BasicStyle
	confirmDialog         *dialog.BasicStyle
	showAcknowledgeDialog = button.Elevated()
	showConfirmDialog     = button.Elevated()
	cartIcon              = wdk.RequireIconWidget(icons.ActionShoppingCart)
	deleteIcon            = wdk.RequireIconWidget(icons.ActionDelete)
)

func appRun(appWindow *app.Window) error {
	var ops op.Ops
	var materialTheme *token.Theme
	for {
		switch windowEvent := appWindow.Event().(type) {
		case app.DestroyEvent:
			return windowEvent.Err

		case app.FrameEvent:
			gtx := app.NewContext(&ops, windowEvent)
			gtx.Values = make(map[string]any)
			if materialTheme == nil {
				scheme := schemes.SchemeBaselineLight()
				materialTheme = defaults.NewTheme(gtx, scheme)
			}
			wdk.InitMaterialThemeInContext(gtx, materialTheme)

			appUpdate(gtx)
			appLayout(gtx)

			windowEvent.Frame(gtx.Ops)
		}
	}
}

func appUpdate(gtx layout.Context) {
	if showAcknowledgeDialog.Clicked(gtx) {
		log.Println("Update cart!")
		acknowledgeDialog = dialog.Acknowledge(gtx)
		acknowledgeDialog.Headline = "Out of stock"
		acknowledgeDialog.Label = "The item in your cart is no longer available."
		appOverlay.Show(acknowledgeDialog.AsOverlayItem())
	}
	if showConfirmDialog.Clicked(gtx) {
		log.Println("Delete messages!")
		confirmDialog = dialog.Confirm(gtx).WithIcon(deleteIcon)
		confirmDialog.Headline = "Permanently Delete?"
		confirmDialog.Label = "Deleting the selected messages will also remove them from synced devices."
		appOverlay.Show(confirmDialog.AsOverlayItem())
	}
	if acknowledgeDialog != nil {
		if acknowledgeDialog.ConfirmButton.Clicked(gtx) {
			log.Println("OK!")
			appOverlay.ClearItem(acknowledgeDialog.GetLayoutItemId())
			acknowledgeDialog = nil
		}
	}
	if confirmDialog != nil {
		dialogButtonClicked := false
		if confirmDialog.ConfirmButton.Clicked(gtx) {
			log.Println("Confirm!")
			dialogButtonClicked = true
		}
		if confirmDialog.CancelButton.Clicked(gtx) {
			log.Println("Cancel!")
			dialogButtonClicked = true
		}
		if dialogButtonClicked {
			appOverlay.ClearItem(confirmDialog.GetLayoutItemId())
			confirmDialog = nil
		}
	}
	appOverlay.Update(gtx)
}

func appLayout(gtx layout.Context) {
	exp.Background(gtx)
	layoutForeground(gtx)
	appOverlay.Layout(gtx)
}

func layoutForeground(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Dialogs!"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return showAcknowledgeDialog.LayoutWithIcon(gtx, "Update cart", cartIcon)
			}),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return showConfirmDialog.LayoutWithIcon(gtx, "Delete messages", deleteIcon)
			}),
		)
	})
}
