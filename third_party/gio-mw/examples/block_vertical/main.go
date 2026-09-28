// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/exp"
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"image"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var appLocale *system.Locale
var appToggleLocale *button.Button

var (
	enLocale     = system.Locale{Language: "en", Direction: system.LTR}
	yiLocale     = system.Locale{Language: "yi", Direction: system.RTL}
	languageIcon = wdk.RequireIconWidget(icons.ActionLanguage)
	lockOpenIcon = wdk.RequireIconWidget(icons.ActionLockOpen)
)

func main() {
	appLocale = &enLocale
	appToggleLocale = button.Elevated()
	go func() {
		var appWindow app.Window
		appWindow.Option(
			app.Title("Block: Vertical"),
			app.Size(unit.Dp(420), unit.Dp(640)),
		)
		if err := appRun(&appWindow); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()
}

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
			gtx.Locale = *appLocale
			appLayout(gtx)

			windowEvent.Frame(gtx.Ops)
		}
	}
}

func appUpdate(gtx layout.Context) {
	if appToggleLocale.Clicked(gtx) {
		if appLocale == &enLocale {
			appLocale = &yiLocale
		} else {
			appLocale = &enLocale
		}
	}
}

func appLayout(gtx layout.Context) {
	exp.Background(gtx)
	layoutForeground(gtx)
}

func layoutForeground(gtx layout.Context) layout.Dimensions {
	btnDims := block.UniformPadding(16).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		label := "Toggle Locale"
		if appLocale == &yiLocale {
			label = "טאָגל שפּראַך"
		}
		return appToggleLocale.LayoutWithIcon(gtx, label, languageIcon)
	})

	// Reset full-viewport min constraints.
	gtx.Constraints.Min.Y = 0
	gtx.Constraints.Min.X = 0
	// Subtract used viewport.
	gtx.Constraints.Max.Y -= btnDims.Size.Y

	tOffset := image.Point{X: 0, Y: btnDims.Size.Y}
	transformStack := op.Offset(tOffset).Push(gtx.Ops)
	dimensions := layoutLineOne(gtx)
	transformStack.Pop()

	return dimensions
}

func layoutLineOne(gtx layout.Context) layout.Dimensions {
	segments := []block.Segment{
		{Widget: layoutL1E1},
		{Widget: layoutL1E2, CrossAlign: block.AlignMiddle},
		{Widget: layoutL1E3},
		{Widget: layoutL1E4},
		{Widget: layoutL1E99},
	}
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
		Expand:   true,
	}.Layout(gtx, segments...)
}

func layoutL1E1(gtx layout.Context) layout.Dimensions {
	theme := wdk.GetMaterialTheme(gtx)

	return block.Background{
		Color: theme.Scheme.Primary.Color,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		c := block.Container{
			MinSize: image.Pt(100, 100),
			Gravity: block.GravityBottomEnd,
		}
		return c.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			txt := "End"
			if appLocale == &yiLocale {
				txt = "סוף"
			}
			presentation := wdk.LabelStyle{
				Color: theme.Scheme.Primary.OnColor,
			}
			return wdk.LayoutLabel(gtx, presentation, txt)
		})
	})
}

func layoutL1E2(gtx layout.Context) layout.Dimensions {
	theme := wdk.GetMaterialTheme(gtx)
	dimensions := lockOpenIcon(gtx, theme.Scheme.Error.Color)
	return dimensions
}

func layoutL1E3(gtx layout.Context) layout.Dimensions {
	theme := wdk.GetMaterialTheme(gtx)
	return block.Background{
		Color: theme.Scheme.Tertiary.Color,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Padding{
			Top:    32,
			Bottom: 32,
			End:    32,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			txt := "Start"
			if appLocale == &yiLocale {
				txt = "אָנהייבן"
			}
			presentation := wdk.LabelStyle{
				Color: theme.Scheme.Tertiary.OnColor,
			}
			return wdk.LayoutLabel(gtx, presentation, txt)
		})
	})
}

func layoutL1E4(gtx layout.Context) layout.Dimensions {
	squareSize := image.Pt(80, 80)
	gtx.Constraints.Min = squareSize
	txt := "Text"
	if appLocale == &yiLocale {
		txt = "טעקסט"
	}
	presentation := wdk.LabelStyle{
		Alignment: text.Middle,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func layoutL1E99(gtx layout.Context) layout.Dimensions {
	theme := wdk.GetMaterialTheme(gtx)
	sColor := theme.Scheme.Secondary.Color.AsNRGBA()
	squareSize := image.Pt(12, 32)
	paint.FillShape(gtx.Ops, sColor, clip.Rect{Max: squareSize}.Op())
	dimensions := layout.Dimensions{Size: squareSize}
	return dimensions
}
