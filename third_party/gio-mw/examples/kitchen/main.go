// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/examples/kitchen/app_shell"
	"gio-mw/wdk"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
)

var appLocale *system.Locale

var (
	enLocale = system.Locale{Language: "en", Direction: system.LTR}
	yiLocale = system.Locale{Language: "yi", Direction: system.RTL}
)

func main() {
	go func() {
		var appWindow app.Window
		appWindow.Option(
			app.Title("Material Example"),
			app.Size(unit.Dp(1280), unit.Dp(640)),
		)
		if err := appRun(&appWindow); err != nil {
			log.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	app.Main()
}

func appRun(appWindow *app.Window) error {
	appLocale = &enLocale
	appShell := app_shell.NewAppShell(appWindow)

	var ops op.Ops
	for {
		switch windowEvent := appWindow.Event().(type) {
		case app.DestroyEvent:
			return windowEvent.Err

		case app.ConfigEvent:

			if windowEvent.Config.Mode == app.Fullscreen {
				appShell.AppFullscreen = true
			} else {
				appShell.AppFullscreen = false
			}

		case app.FrameEvent:
			gtx := app.NewContext(&ops, windowEvent)
			if appShell.AppTheme == nil {
				scheme := schemes.SchemeBaselineLight()
				appShell.AppTheme = defaults.NewTheme(gtx, scheme)
			}
			gtx.Values = make(map[string]any)
			wdk.InitMaterialThemeInContext(gtx, appShell.AppTheme)
			gtx.Locale = *appLocale

			area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
			event.Op(gtx.Ops, appWindow)

			for {
				keyEv, ok := gtx.Event(
					key.Filter{Name: key.NameEscape},
					key.Filter{Name: key.NameF11},
				)
				if !ok {
					break
				}
				switch keyEv := keyEv.(type) {
				case key.Event:
					if keyEv.State != key.Release {
						break
					}
					if keyEv.Name == key.NameEscape {
						return nil
					}
					if keyEv.Name == key.NameF11 {
						if appShell.AppFullscreen {
							appWindow.Option(app.Windowed.Option())
						} else {
							appWindow.Option(app.Fullscreen.Option())
						}
					}
				}
			}

			appShell.Update(gtx)
			appShell.Layout(gtx)
			area.Pop()
			windowEvent.Frame(gtx.Ops)
		}
	}
}
