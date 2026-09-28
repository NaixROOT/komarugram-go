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
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"gioui.org/x/debug"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

func main() {
	go func() {
		var appWindow app.Window
		appWindow.Option(
			app.Title("Snackbar Example"),
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
	appOverlay overlay.Overlay

	appButtonIcon = wdk.RequireIconWidget(icons.ActionDone)
	appButton     = button.Elevated()
	debugTag      = true
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
		log.Println("Hi!")
		snackbarStyle := snackbar.Plain("Hi!")
		appOverlay.Show(overlay.NewItem(snackbarStyle.Layout, block.GravityBottomCenter).WithDuration(3 * time.Second))
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
				txt := "Snackbars!"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return debug.Layout(gtx, &debugTag, func(gtx layout.Context) layout.Dimensions {
					txt := "This is a simple body message, click for debug info."
					return exp.BodyL(gtx, txt)
				})
			}),
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return appButton.LayoutWithIcon(gtx, "Say Hi!", appButtonIcon)
			}),
		)
	})
}
