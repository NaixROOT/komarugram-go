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
	"gio-mw/widget/card"
	"gio-mw/widget/overlay"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var appLocale *system.Locale
var appToggleLocale *button.Button
var appToggleElevation *button.Button
var appWindow app.Window
var appOverlay overlay.Overlay
var frameRateCounter *exp.FrameRateCounter

var (
	enLocale        = system.Locale{Language: "en", Direction: system.LTR}
	yiLocale        = system.Locale{Language: "yi", Direction: system.RTL}
	languageIcon    = wdk.RequireIconWidget(icons.ActionLanguage)
	elevationActive = false
)

func main() {
	appLocale = &enLocale
	appToggleLocale = button.Elevated()
	appToggleElevation = button.Elevated()
	frameRateCounter = &exp.FrameRateCounter{}

	go func() {
		appWindow.Option(
			app.Title("Surfaces Example"),
			app.Size(unit.Dp(640), unit.Dp(920)),
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

			// Measure performance from here.
			appUpdate(gtx)
			gtx.Locale = *appLocale
			appLayout(gtx)
			// Measure performance to here.
			gtx.Execute(op.InvalidateCmd{})

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
	if appToggleElevation.Clicked(gtx) {
		elevationActive = !elevationActive
	}
	appOverlay.Update(gtx)
	frameRateCounter.Update(gtx)
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
				frameRateText := "Counting frames..."
				if frameRateCounter.Rate > 0 {
					frameRateText = fmt.Sprintf("Frame Rate: %02.2f", frameRateCounter.Rate)
					if appLocale == &yiLocale {
						frameRateText = fmt.Sprintf("פרעים ראַטע: %02f", frameRateCounter.Rate)
					}
				}
				return exp.BodyS(gtx, frameRateText)
			}),
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
				label := "Toggle Elevation"
				if appLocale == &yiLocale {
					label = "טאָגל הייבונג"
				}
				return appToggleElevation.Layout(gtx, label)
			}),
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				c := card.Card{Kind: card.Elevated}
				return c.Layout(gtx,
					card.Content(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
						return layoutCardContents(gtx)
					}),
				)
			}),
		)
	})
}

func layoutCardContents(gtx layout.Context) layout.Dimensions {
	maxPoint := f32.Point{X: float32(gtx.Constraints.Max.X), Y: float32(gtx.Constraints.Max.Y)}
	rect := wdk.ShapedRect{
		MaxPoint: maxPoint,
		Shapes: wdk.CornerShapes{
			TopEnd: wdk.CornerShape{
				Kind: wdk.CornerKindChamfer,
				Size: float32(gtx.Dp(28)),
			},
		},
	}
	defer exp.PrimaryContainer(gtx, rect.Outline(gtx)).Pop()
	return block.Container{
		MaxSize: maxPoint.Round(),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return exp.HeadlineL(gtx, "Aloha!")
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				levelNumber := 0
				return layoutRecursiveLevel(gtx, levelNumber)
			}),
		)
	})
}

// WARNING: This is called recursively, components should avoid infinite recursion.
func layoutRecursiveLevel(gtx layout.Context, levelNumber int) layout.Dimensions {
	levelNumber++
	return block.UniformPadding(examples.SpacingSmall).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		maxPoint := gtx.Constraints.Max
		materialTheme := wdk.GetMaterialTheme(gtx)
		cardTheme := wdk.GetWidgetTheme[card.Theme](materialTheme, card.ElevatedNamespace)
		baseBox := wdk.Box{
			EndPoint: maxPoint,
			Shape:    wdk.FromCornerShapesToken(gtx, cardTheme.EnabledContainerShape),
		}

		if elevationActive {
			lElevation := wdk.Elevation{
				Level:       token.ElevationLevel1,
				ShadowColor: materialTheme.Scheme.Shadow,
			}
			lElevation.Layout(gtx, baseBox)
		}

		colorChoice := levelNumber % 4
		switch colorChoice {
		case 0:
			defer exp.PrimaryContainer(gtx, baseBox.Outline(gtx)).Pop()
		case 1:
			defer exp.SecondaryContainer(gtx, baseBox.Outline(gtx)).Pop()
		case 2:
			defer exp.TertiaryContainer(gtx, baseBox.Outline(gtx)).Pop()
		case 3:
			defer exp.InverseSurface(gtx, baseBox.Outline(gtx)).Pop()
		default:
			defer exp.InverseSurface(gtx, baseBox.Outline(gtx)).Pop()
		}

		return block.Container{
			MinSize: maxPoint,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return block.Line{
					Axis:     block.AxisVertical,
					Overflow: block.OverflowClip,
				}.Layout(gtx,
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						return exp.BodyS(gtx, fmt.Sprintf("Level %d: %d", levelNumber, colorChoice))
					}),
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						// WARNING: This is called recursively, components should avoid infinite recursion.
						return layoutRecursiveLevel(gtx, levelNumber)
					}),
				)
			})
		})
	})
}
