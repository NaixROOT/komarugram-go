// SPDX-License-Identifier: Unlicense OR MIT

package search

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.search.Search"

const (
	iconSize          = unit.Dp(24)
	paddingBottom     = unit.Dp(8)
	paddingHorizontal = unit.Dp(16)
	paddingTop        = unit.Dp(8)
	widthMin          = unit.Dp(240)
	widthMax          = unit.Dp(720)
)

type Theme struct {
	EnabledContainerColor     token.MatColor
	EnabledContainerElevation token.ElevationLevel
	EnabledContainerHeight    unit.Dp
	EnabledContainerShape     token.CornerShapes
	//EnabledAvatarShape         token.CornerShapes
	//EnabledAvatarSize          unit.Dp
	EnabledLeadingIconColor    token.MatColor
	EnabledTrailingIconColor   token.MatColor
	EnabledSupportingTextColor token.MatColor
	//EnabledSupportingTextType  token.TypeInfo
	EnabledInputTextColor token.MatColor
	//EnabledInputTextType  token.TypeInfo

	HoveredStateLayerColor     token.MatColor
	HoveredStateLayerOpacity   token.OpacityLevel
	HoveredSupportingTextColor token.MatColor

	FocusedIndicatorColor     token.MatColor
	FocusedIndicatorThickness unit.Dp
	FocusedIndicatorOffset    unit.Dp

	PressedStateLayerColor     token.MatColor
	PressedStateLayerOpacity   token.OpacityLevel
	PressedSupportingTextColor token.MatColor
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerColor:     materialTheme.Scheme.SurfaceContainerHigh,
		EnabledContainerElevation: token.ElevationLevel3,
		EnabledContainerHeight:    unit.Dp(56),
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		EnabledLeadingIconColor:    materialTheme.Scheme.Surface.OnColor,
		EnabledTrailingIconColor:   materialTheme.Scheme.SurfaceVariant.OnColor,
		EnabledSupportingTextColor: materialTheme.Scheme.SurfaceVariant.OnColor,
		//EnabledSupportingTextType:  materialTheme.Typescale.BodySmall,
		EnabledInputTextColor: materialTheme.Scheme.Surface.OnColor,
		//EnabledInputTextType:  materialTheme.Typescale.BodyLarge,

		HoveredStateLayerColor:     materialTheme.Scheme.Surface.OnColor,
		HoveredStateLayerOpacity:   token.OpacityLevel2,
		HoveredSupportingTextColor: materialTheme.Scheme.SurfaceVariant.OnColor,

		FocusedIndicatorColor:     materialTheme.Scheme.Surface.OnColor,
		FocusedIndicatorOffset:    unit.Dp(3),
		FocusedIndicatorThickness: unit.Dp(2),

		PressedStateLayerColor:     materialTheme.Scheme.Surface.OnColor,
		PressedStateLayerOpacity:   token.OpacityLevel3,
		PressedSupportingTextColor: materialTheme.Scheme.SurfaceVariant.OnColor,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
