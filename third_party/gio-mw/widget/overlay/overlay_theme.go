// SPDX-License-Identifier: Unlicense OR MIT

package overlay

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
)

const Namespace = "gio-mw.widget.overlay.Overlay"

type Theme struct {
	EnabledScrimColor   token.MatColor
	EnabledScrimOpacity token.OpacityLevel
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledScrimColor:   materialTheme.Scheme.Scrim,
		EnabledScrimOpacity: token.OpacityLevel8,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
