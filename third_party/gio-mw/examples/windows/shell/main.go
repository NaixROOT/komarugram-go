// SPDX-License-Identifier: Unlicense OR MIT

package shell

import (
	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

const themeKey = "gio-mw.theme"

type MainShell struct {
	Application   *Application
	Window        *app.Window
	materialTheme *token.Theme
}

func NewMainShell(application *Application) {
	var shellWindow app.Window
	shellWindow.Option(
		app.Title("Main Window"),
		app.Size(unit.Dp(480), unit.Dp(480)),
	)
	shell := MainShell{
		Application: application,
		Window:      &shellWindow,
	}
	shell.Application.active.Add(1)

	go func() {
		defer shell.Application.active.Done()

		err := shell.Run()
		if err != nil {
			shell.Application.Shutdown()
		}
	}()
}

func (s *MainShell) Run() error {
	go func() {
		<-s.Application.Context.Done()
		s.Window.Perform(system.ActionClose)
	}()

	var ops op.Ops
	for {
		switch windowEvent := s.Window.Event().(type) {
		case app.DestroyEvent:
			return windowEvent.Err

		case app.FrameEvent:
			gtx := app.NewContext(&ops, windowEvent)
			gtx.Values = make(map[string]any)
			if s.materialTheme == nil {
				scheme := schemes.SchemeBaselineLight()
				s.materialTheme = defaults.NewTheme(gtx, scheme)
			}
			wdk.InitMaterialThemeInContext(gtx, s.materialTheme)
			s.update(gtx)
			gtx.Locale = s.Application.Locale
			s.layout(gtx)

			windowEvent.Frame(gtx.Ops)
		}
	}
}

func (s *MainShell) update(gtx layout.Context) {
}

func (s *MainShell) layout(gtx layout.Context) {
	block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
		Expand:   true,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			presentation := wdk.LabelStyle{
				Typestyle: token.TypestyleHeadlineSmall,
			}
			return wdk.LayoutLabel(gtx, presentation, "Hello World")
		}),
	)
}
