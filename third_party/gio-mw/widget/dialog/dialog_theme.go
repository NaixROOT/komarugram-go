// SPDX-License-Identifier: Unlicense OR MIT

package dialog

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.dialog"
const BasicNamespace = "gio-mw.widget.dialog.Basic"
const FullScreenNamespace = "gio-mw.widget.dialog.FullScreen"

type Theme struct {
	// TODO: Add button label font style properties.
	EnabledContainerColor       token.MatColor
	EnabledContainerElevation   token.ElevationLevel
	EnabledContainerShadowColor token.MatColor
	EnabledContainerShape       token.CornerShapes
	EnabledLabelColor           token.MatColor
	EnabledIconColor            token.MatColor
	EnabledIconSize             unit.Dp
	EnabledSubheadColor         token.MatColor
	EnabledHeadlineColor        token.MatColor
	EnabledDividerColor         token.MatColor
	EnabledDividerSize          unit.Dp
	EnabledSupportingTextColor  token.MatColor
}

func BuildBasicTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, BasicNamespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerColor:       materialTheme.Scheme.SurfaceContainerHigh,
		EnabledContainerElevation:   token.ElevationLevel3,
		EnabledContainerShadowColor: materialTheme.Scheme.Shadow,
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(28),
		}),
		EnabledLabelColor:          materialTheme.Scheme.SurfaceVariant.OnColor,
		EnabledIconColor:           materialTheme.Scheme.Secondary.Color,
		EnabledIconSize:            unit.Dp(24),
		EnabledSubheadColor:        materialTheme.Scheme.Surface.OnColor,
		EnabledHeadlineColor:       materialTheme.Scheme.Surface.OnColor,
		EnabledDividerColor:        materialTheme.Scheme.Outline,
		EnabledDividerSize:         unit.Dp(1),
		EnabledSupportingTextColor: materialTheme.Scheme.SurfaceVariant.OnColor,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, BasicNamespace, widgetTheme)
	return widgetTheme
}
