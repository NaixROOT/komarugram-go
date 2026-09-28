// SPDX-License-Identifier: Unlicense OR MIT

package divider

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
)

const Namespace = "gio-mw.widget.divider.Divider"

type Theme struct {
	EnabledContainerColor token.MatColor
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerColor: materialTheme.Scheme.OutlineVariant,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
