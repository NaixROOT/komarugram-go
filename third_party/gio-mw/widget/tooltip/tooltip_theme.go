// SPDX-License-Identifier: Unlicense OR MIT

package tooltip

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.tooltip.Tooltip"

type Theme struct {
	EnabledContainerColor      token.MatColor
	EnabledContainerShape      token.CornerShapes
	EnabledSupportingTextColor token.MatColor

	// Extra configuration options:
	paddingContent unit.Dp
	maxWidth       unit.Dp
	minHeight      unit.Dp
	targetSpacing  unit.Dp
}

func BuildPlainTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerColor: materialTheme.Scheme.InverseSurface.Color,
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(4),
		}),
		EnabledSupportingTextColor: materialTheme.Scheme.InverseSurface.OnColor,

		minHeight:      unit.Dp(24),
		maxWidth:       unit.Dp(320),
		paddingContent: unit.Dp(8),
		targetSpacing:  unit.Dp(4),
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
