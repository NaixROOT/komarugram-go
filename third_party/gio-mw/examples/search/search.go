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
	"gio-mw/widget/overlay"
	"gio-mw/widget/search"
	"gio-mw/widget/snackbar"
	"log"
	"os"
	"time"

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
			app.Title("Search Example"),
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
	appOverlay   overlay.Overlay
	appSearchBar search.Search
)

func appRun(appWindow *app.Window) error {
	appSearchBar = *search.Bar()
	appSearchBar.SupportingText = "Search content"
	appSearchBar.LeadingIcon.Icon = wdk.RequireIconWidget(icons.ActionSearch)
	appSearchBar.TrailingIcon.Icon = wdk.RequireIconWidget(icons.ContentClear)
	appSearchBar.TrailingIcon.Label = "Clear"

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
	if appSearchBar.Submitted(gtx) {
		searchQuery := appSearchBar.GetText()
		snackbarStyle := snackbar.Plain(fmt.Sprintf("Searching for '%s'", searchQuery))
		appOverlay.Show(overlay.NewItem(snackbarStyle.Layout, block.GravityBottomCenter).WithDuration(3 * time.Second))
	}
	if appSearchBar.TrailingIcon.Clickable.Clicked(gtx) {
		appSearchBar.ClearText()
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
				txt := "Search!"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Search lets people enter a keyword or phrase to get relevant information."
				return exp.BodyL(gtx, txt)
			}),
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return appSearchBar.Layout(gtx)
			}),
		)
	})
}
