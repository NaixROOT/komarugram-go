// SPDX-License-Identifier: Unlicense OR MIT

package snackbar

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.snackbar.Snackbar"

type Theme struct {
	EnabledContainerColor         token.MatColor
	EnabledContainerElevation     token.ElevationLevel
	EnabledContainerShadowColor   token.MatColor
	EnabledContainerShape         token.CornerShapes
	EnabledContainerOneLineHeight unit.Dp
	EnabledContainerTwoLineHeight unit.Dp
	EnabledLabelColor             token.MatColor
	EnabledIconColor              token.MatColor
	EnabledIconSize               unit.Dp
	EnabledSupportingTextColor    token.MatColor

	HoveredStateLayerColor   token.MatColor
	HoveredStateLayerOpacity token.OpacityLevel
	HoveredLabelColor        token.MatColor
	HoveredIconColor         token.MatColor

	FocusedStateLayerColor   token.MatColor
	FocusedStateLayerOpacity token.OpacityLevel
	FocusedLabelColor        token.MatColor
	FocusedIconColor         token.MatColor

	PressedStateLayerColor   token.MatColor
	PressedStateLayerOpacity token.OpacityLevel
	PressedLabelColor        token.MatColor
	PressedIconColor         token.MatColor
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerColor:       materialTheme.Scheme.InverseSurface.Color,
		EnabledContainerElevation:   token.ElevationLevel3,
		EnabledContainerShadowColor: materialTheme.Scheme.Shadow,
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(4),
		}),
		EnabledContainerOneLineHeight: unit.Dp(48),
		EnabledContainerTwoLineHeight: unit.Dp(68),
		EnabledLabelColor:             materialTheme.Scheme.InversePrimary,
		EnabledIconColor:              materialTheme.Scheme.InverseSurface.OnColor,
		EnabledIconSize:               unit.Dp(24),
		EnabledSupportingTextColor:    materialTheme.Scheme.InverseSurface.OnColor,
		HoveredStateLayerColor:        materialTheme.Scheme.InversePrimary,
		HoveredStateLayerOpacity:      token.OpacityLevel2,
		HoveredLabelColor:             materialTheme.Scheme.InversePrimary,
		HoveredIconColor:              materialTheme.Scheme.InverseSurface.OnColor,
		FocusedStateLayerColor:        materialTheme.Scheme.InversePrimary,
		FocusedStateLayerOpacity:      token.OpacityLevel3,
		FocusedLabelColor:             materialTheme.Scheme.InversePrimary,
		FocusedIconColor:              materialTheme.Scheme.InverseSurface.OnColor,
		PressedStateLayerColor:        materialTheme.Scheme.InversePrimary,
		PressedStateLayerOpacity:      token.OpacityLevel3,
		PressedLabelColor:             materialTheme.Scheme.InversePrimary,
		PressedIconColor:              materialTheme.Scheme.InverseSurface.OnColor,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
