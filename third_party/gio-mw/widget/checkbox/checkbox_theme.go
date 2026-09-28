// SPDX-License-Identifier: Unlicense OR MIT

package checkbox

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.checkbox.Checkboxes"

const (
	iconStrokeWidth = unit.Dp(2)
)

type Theme struct {
	EnabledContainerSelectedColor      token.MatColor
	EnabledContainerSelectedErrorColor token.MatColor
	EnabledContainerShape              token.CornerShapes
	EnabledContainerSize               unit.Dp
	EnabledOutlineUnselectedColor      token.MatColor
	EnabledOutlineUnselectedErrorColor token.MatColor
	EnabledOutlineUnselectedWidth      unit.Dp
	EnabledIconSize                    unit.Dp
	EnabledIconSelectedColor           token.MatColor
	EnabledIconSelectedErrorColor      token.MatColor
	EnabledIconUnselectedColor         token.MatColor
	EnabledIconUnselectedErrorColor    token.MatColor
	EnabledStateLayerShape             token.CornerShapes
	EnabledStateLayerSize              unit.Dp

	DisabledContainerSelectedColor   token.MatColor
	DisabledContainerSelectedOpacity token.OpacityLevel
	DisabledOutlineUnselectedColor   token.MatColor
	DisabledOutlineUnselectedOpacity token.OpacityLevel
	DisabledIconSelectedColor        token.MatColor
	DisabledIconSelectedOpacity      token.OpacityLevel
	DisabledIconUnselectedColor      token.MatColor
	DisabledIconUnselectedOpacity    token.OpacityLevel

	HoveredContainerSelectedColor      token.MatColor
	HoveredContainerSelectedErrorColor token.MatColor
	HoveredOutlineUnselectedColor      token.MatColor
	HoveredOutlineUnselectedErrorColor token.MatColor
	HoveredStateLayerSelectedColor     token.MatColor
	HoveredStateLayerSelectedOpacity   token.OpacityLevel
	HoveredStateLayerUnselectedColor   token.MatColor
	HoveredStateLayerUnselectedOpacity token.OpacityLevel
	HoveredStateLayerErrorColor        token.MatColor
	HoveredStateLayerErrorOpacity      token.OpacityLevel
	HoveredIconSelectedColor           token.MatColor
	HoveredIconUnselectedColor         token.MatColor
	HoveredIconErrorColor              token.MatColor

	FocusedFocusIndicatorColor         token.MatColor
	FocusedFocusIndicatorThickness     unit.Dp
	FocusedFocusIndicatorOffset        unit.Dp
	FocusedContainerSelectedColor      token.MatColor
	FocusedContainerSelectedErrorColor token.MatColor
	FocusedOutlineUnselectedColor      token.MatColor
	FocusedOutlineUnselectedErrorColor token.MatColor
	FocusedStateLayerSelectedColor     token.MatColor
	FocusedStateLayerSelectedOpacity   token.OpacityLevel
	FocusedStateLayerUnselectedColor   token.MatColor
	FocusedStateLayerUnselectedOpacity token.OpacityLevel
	FocusedStateLayerErrorColor        token.MatColor
	FocusedStateLayerErrorOpacity      token.OpacityLevel
	FocusedIconSelectedColor           token.MatColor
	FocusedIconUnselectedColor         token.MatColor
	FocusedIconErrorColor              token.MatColor

	PressedContainerSelectedColor      token.MatColor
	PressedContainerSelectedErrorColor token.MatColor
	PressedOutlineUnselectedColor      token.MatColor
	PressedOutlineUnselectedErrorColor token.MatColor
	PressedStateLayerSelectedColor     token.MatColor
	PressedStateLayerSelectedOpacity   token.OpacityLevel
	PressedStateLayerUnselectedColor   token.MatColor
	PressedStateLayerUnselectedOpacity token.OpacityLevel
	PressedStateLayerErrorColor        token.MatColor
	PressedStateLayerErrorOpacity      token.OpacityLevel
	PressedIconSelectedColor           token.MatColor
	PressedIconUnselectedColor         token.MatColor
	PressedIconErrorColor              token.MatColor
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerSelectedColor:      materialTheme.Scheme.Primary.Color,
		EnabledContainerSelectedErrorColor: materialTheme.Scheme.Error.Color,
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(2),
		}),
		EnabledContainerSize:               unit.Dp(18),
		EnabledOutlineUnselectedColor:      materialTheme.Scheme.SurfaceVariant.OnColor,
		EnabledOutlineUnselectedErrorColor: materialTheme.Scheme.Error.Color,
		EnabledOutlineUnselectedWidth:      unit.Dp(2),
		EnabledIconSize:                    unit.Dp(18),
		EnabledIconSelectedColor:           materialTheme.Scheme.Primary.OnColor,
		EnabledIconSelectedErrorColor:      materialTheme.Scheme.Error.OnColor,
		EnabledIconUnselectedColor:         materialTheme.Scheme.Surface.Color,
		EnabledIconUnselectedErrorColor:    materialTheme.Scheme.Error.OnColor,
		EnabledStateLayerShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		EnabledStateLayerSize: unit.Dp(40),

		DisabledContainerSelectedColor:   materialTheme.Scheme.Surface.OnColor,
		DisabledContainerSelectedOpacity: token.OpacityLevel9,
		DisabledOutlineUnselectedColor:   materialTheme.Scheme.Surface.OnColor,
		DisabledOutlineUnselectedOpacity: token.OpacityLevel9,
		DisabledIconSelectedColor:        materialTheme.Scheme.Primary.Color,
		DisabledIconSelectedOpacity:      token.OpacityLevel9,
		DisabledIconUnselectedColor:      materialTheme.Scheme.Primary.Color,
		DisabledIconUnselectedOpacity:    token.OpacityLevel9,

		HoveredContainerSelectedColor:      materialTheme.Scheme.Primary.Color,
		HoveredContainerSelectedErrorColor: materialTheme.Scheme.Error.Color,
		HoveredOutlineUnselectedColor:      materialTheme.Scheme.Primary.Color,
		HoveredOutlineUnselectedErrorColor: materialTheme.Scheme.Error.Color,
		HoveredStateLayerSelectedColor:     materialTheme.Scheme.Primary.Color,
		HoveredStateLayerSelectedOpacity:   token.OpacityLevel2,
		HoveredStateLayerUnselectedColor:   materialTheme.Scheme.Surface.OnColor,
		HoveredStateLayerUnselectedOpacity: token.OpacityLevel2,
		HoveredStateLayerErrorColor:        materialTheme.Scheme.Error.Color,
		HoveredStateLayerErrorOpacity:      token.OpacityLevel2,
		HoveredIconSelectedColor:           materialTheme.Scheme.Primary.OnColor,
		HoveredIconUnselectedColor:         materialTheme.Scheme.Surface.OnColor,
		HoveredIconErrorColor:              materialTheme.Scheme.Error.OnColor,

		FocusedFocusIndicatorColor:         materialTheme.Scheme.Secondary.Color,
		FocusedFocusIndicatorThickness:     unit.Dp(3),
		FocusedFocusIndicatorOffset:        unit.Dp(2),
		FocusedContainerSelectedColor:      materialTheme.Scheme.Primary.Color,
		FocusedContainerSelectedErrorColor: materialTheme.Scheme.Error.Color,
		FocusedOutlineUnselectedColor:      materialTheme.Scheme.Surface.OnColor,
		FocusedOutlineUnselectedErrorColor: materialTheme.Scheme.Error.Color,
		FocusedStateLayerSelectedColor:     materialTheme.Scheme.Primary.Color,
		FocusedStateLayerSelectedOpacity:   token.OpacityLevel3,
		FocusedStateLayerUnselectedColor:   materialTheme.Scheme.Surface.OnColor,
		FocusedStateLayerUnselectedOpacity: token.OpacityLevel3,
		FocusedStateLayerErrorColor:        materialTheme.Scheme.Error.Color,
		FocusedStateLayerErrorOpacity:      token.OpacityLevel3,
		FocusedIconSelectedColor:           materialTheme.Scheme.Primary.OnColor,
		FocusedIconUnselectedColor:         materialTheme.Scheme.Surface.OnColor,
		FocusedIconErrorColor:              materialTheme.Scheme.Error.OnColor,

		PressedContainerSelectedColor:      materialTheme.Scheme.Primary.Color,
		PressedContainerSelectedErrorColor: materialTheme.Scheme.Error.Color,
		PressedOutlineUnselectedColor:      materialTheme.Scheme.Primary.Color,
		PressedOutlineUnselectedErrorColor: materialTheme.Scheme.Error.Color,
		PressedStateLayerSelectedColor:     materialTheme.Scheme.Surface.Color,
		PressedStateLayerSelectedOpacity:   token.OpacityLevel3,
		PressedStateLayerUnselectedColor:   materialTheme.Scheme.Primary.OnColor,
		PressedStateLayerUnselectedOpacity: token.OpacityLevel3,
		PressedStateLayerErrorColor:        materialTheme.Scheme.Error.Color,
		PressedStateLayerErrorOpacity:      token.OpacityLevel3,
		PressedIconSelectedColor:           materialTheme.Scheme.Primary.OnColor,
		PressedIconUnselectedColor:         materialTheme.Scheme.Surface.OnColor,
		PressedIconErrorColor:              materialTheme.Scheme.Error.OnColor,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
