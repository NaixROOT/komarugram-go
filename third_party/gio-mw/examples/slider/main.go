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
	"gio-mw/widget/slider"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var appLocale *system.Locale
var appToggleLocale *button.Button

var (
	enLocale     = system.Locale{Language: "en", Direction: system.LTR}
	yiLocale     = system.Locale{Language: "yi", Direction: system.RTL}
	languageIcon = wdk.RequireIconWidget(icons.ActionLanguage)
)

func main() {
	appLocale = &enLocale
	appToggleLocale = button.Elevated()
	go func() {
		var appWindow app.Window
		appWindow.Option(
			app.Title("Slider Example"),
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
	appOverlay         overlay.Overlay
	standardSlider     *slider.Slider
	centeredSlider     *slider.Slider
	rangeSlider        *slider.Slider
	incrementSliderBtn *button.Button
	decrementSliderBtn *button.Button
	volumeDownIcon     = wdk.RequireIconWidget(icons.AVVolumeDown)
	volumeUpIcon       = wdk.RequireIconWidget(icons.AVVolumeUp)
)

func appRun(appWindow *app.Window) error {
	var ops op.Ops
	var materialTheme *token.Theme
	standardSliderOptions := make([]int, 11)
	for i := 0; i < len(standardSliderOptions); i++ {
		standardSliderOptions[i] = i
	}
	standardSlider = slider.StandardSlider(standardSliderOptions, 5, func(value int) {
		log.Printf("Standard Slider Value: %d", value)
	})
	incrementSliderBtn = button.Text()
	decrementSliderBtn = button.Text()
	centeredSliderOptions := make([]int, 21)
	sliderSideSize := len(centeredSliderOptions) / 2
	for i := -sliderSideSize; i < sliderSideSize; i++ {
		centeredSliderOptions[i+10] = i
	}
	centeredSlider = slider.CenteredSlider(centeredSliderOptions, 0, func(value int) {
		log.Printf("Centered Slider Value: %d", value)
	})
	rangeSliderOptions := make([]int, 100)
	for i := 0; i < 100; i++ {
		rangeSliderOptions[i] = i
	}
	rangeSlider = slider.RangeSlider(rangeSliderOptions, 50, 75)
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
	if incrementSliderBtn.Clicked(gtx) && standardSlider.GetValue() < 10 {
		standardSlider.SetValue(standardSlider.GetValue() + 1)
	}
	if decrementSliderBtn.Clicked(gtx) && standardSlider.GetValue() > 0 {
		standardSlider.SetValue(standardSlider.GetValue() - 1)
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
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				label := "Hello World!"
				if appLocale == &yiLocale {
					label = "העלא, וועלט!"
				}
				presentation := wdk.LabelStyle{
					Typestyle: token.TypestyleLabelLarge,
				}
				return wdk.LayoutLabel(gtx, presentation, label)
			}),
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				label := "Toggle Locale"
				if appLocale == &yiLocale {
					label = "טאָגל שפּראַך"
				}
				return appToggleLocale.LayoutWithIcon(gtx, label, languageIcon)
			}),
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = 0
				return block.Line{
					Axis:     block.AxisHorizontal,
					Overflow: block.OverflowClip,
					Expand:   false,
				}.Layout(gtx,
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						return decrementSliderBtn.LayoutIconOnly(gtx, "Decrement", volumeDownIcon)
					}),
					block.NewHorizontalSpacer(examples.SpacingMedium),
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						return incrementSliderBtn.LayoutIconOnly(gtx, "Increment", volumeUpIcon)
					}),
				)
			}).AlignMiddle(),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return standardSlider.Layout(gtx)
			}),
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return centeredSlider.Layout(gtx)
			}),
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return rangeSlider.Layout(gtx)
			}),
		)
	})
}
