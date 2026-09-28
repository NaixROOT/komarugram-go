// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"fmt"
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
	"gio-mw/widget/sheet"
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
			app.Title("Sheet Example"),
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
	appOverlay        overlay.Overlay
	acknowledgeDialog *dialog.BasicStyle
	confirmDialog     *dialog.BasicStyle
	openBottomSheet   = button.Elevated()
	openSideSheet     = button.Elevated()
	bottomIcon        = wdk.RequireIconWidget(icons.EditorBorderBottom)
	sideIcon          = wdk.RequireIconWidget(icons.EditorBorderRight)
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
	if openBottomSheet.Clicked(gtx) {
		log.Println("Open bottom sheet!")
		appOverlay.Show(sheet.NewSheet(sheet.Bottom, layoutSheetContents))
	}
	if openSideSheet.Clicked(gtx) {
		log.Println("Open end sheet!")
		appOverlay.Show(sheet.NewSheet(sheet.Side, layoutSheetContents))
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
				txt := "Sheets!"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return openBottomSheet.LayoutWithIcon(gtx, "Open bottom sheet", bottomIcon)
			}),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return openSideSheet.LayoutWithIcon(gtx, "Open side sheet", sideIcon)
			}),
		)
	})
}

func layoutSheetContents(gtx layout.Context) layout.Dimensions {
	return block.Padding{
		Start: examples.SpacingMedium,
		End:   examples.SpacingMedium,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		segments := make([]block.Segment, 0)
		for i := 0; i < 40; i++ {
			segments = append(segments, block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := fmt.Sprintf("Hello World #%d!", i)
				return exp.HeadlineL(gtx, txt)
			}))
		}
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx, segments...)
	})
}
