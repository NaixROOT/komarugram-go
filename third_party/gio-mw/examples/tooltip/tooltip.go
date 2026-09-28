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
	"gio-mw/widget/overlay"
	"gio-mw/widget/snackbar"
	"gio-mw/widget/tooltip"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

func main() {
	go func() {
		var appWindow app.Window
		appWindow.Option(
			app.Title("Tooltip Example"),
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
	appOverlay              overlay.Overlay
	appButton               = button.Elevated()
	appButtonDisplayTooltip = true
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
	if appButton.Clicked(gtx) {
		snackbarStyle := snackbar.Plain("Hi!")
		appOverlay.Show(overlay.NewItem(snackbarStyle.Layout, block.GravityBottomCenter))
	}
	if appButton.Hovered(gtx) {
		appButtonDisplayTooltip = true
	} else {
		appButtonDisplayTooltip = false
	}
	appOverlay.Update(gtx)
}

func appLayout(gtx layout.Context) {
	exp.Background(gtx)
	layoutForeground(gtx)
	appOverlay.Layout(gtx)
}

func layoutForeground(gtx layout.Context) layout.Dimensions {
	return layout.UniformInset(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Tooltips!"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Tooltips add additional context to UI elements, they are displayed on hover."
				return exp.BodyL(gtx, txt)
			}),
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				t := tooltip.PlainTooltip("Click this strange button to see what it does.")
				return t.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return appButton.Layout(gtx, "Strange button")
				}, appButtonDisplayTooltip)
			}),
		)
	})
}
