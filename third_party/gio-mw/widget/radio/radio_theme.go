// SPDX-License-Identifier: Unlicense OR MIT

package radio

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.radio.Radios"

const (
	outlineStrokeWidth      = unit.Dp(2)
	focusOutlineStrokeWidth = unit.Dp(3)
)

type Theme struct {
	EnabledIconSize            unit.Dp
	EnabledIconSelectedColor   token.MatColor
	EnabledIconUnselectedColor token.MatColor
	EnabledStateLayerShape     token.CornerShapes
	EnabledStateLayerSize      unit.Dp

	DisabledIconSelectedColor     token.MatColor
	DisabledIconSelectedOpacity   token.OpacityLevel
	DisabledIconUnselectedColor   token.MatColor
	DisabledIconUnselectedOpacity token.OpacityLevel

	HoveredStateLayerSelectedColor     token.MatColor
	HoveredStateLayerSelectedOpacity   token.OpacityLevel
	HoveredStateLayerUnselectedColor   token.MatColor
	HoveredStateLayerUnselectedOpacity token.OpacityLevel
	HoveredIconSelectedColor           token.MatColor
	HoveredIconUnselectedColor         token.MatColor

	FocusedStateLayerSelectedColor     token.MatColor
	FocusedStateLayerSelectedOpacity   token.OpacityLevel
	FocusedStateLayerUnselectedColor   token.MatColor
	FocusedStateLayerUnselectedOpacity token.OpacityLevel
	FocusedIconSelectedColor           token.MatColor
	FocusedIconUnselectedColor         token.MatColor

	PressedStateLayerSelectedColor     token.MatColor
	PressedStateLayerSelectedOpacity   token.OpacityLevel
	PressedStateLayerUnselectedColor   token.MatColor
	PressedStateLayerUnselectedOpacity token.OpacityLevel
	PressedIconSelectedColor           token.MatColor
	PressedIconUnselectedColor         token.MatColor
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledIconSize:            unit.Dp(20),
		EnabledIconSelectedColor:   materialTheme.Scheme.Primary.Color,
		EnabledIconUnselectedColor: materialTheme.Scheme.SurfaceVariant.OnColor,
		EnabledStateLayerShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		EnabledStateLayerSize: unit.Dp(40),

		DisabledIconSelectedColor:     materialTheme.Scheme.Primary.Color,
		DisabledIconSelectedOpacity:   token.OpacityLevel9,
		DisabledIconUnselectedColor:   materialTheme.Scheme.Surface.OnColor,
		DisabledIconUnselectedOpacity: token.OpacityLevel9,

		HoveredStateLayerSelectedColor:     materialTheme.Scheme.Primary.Color,
		HoveredStateLayerSelectedOpacity:   token.OpacityLevel2,
		HoveredStateLayerUnselectedColor:   materialTheme.Scheme.Surface.OnColor,
		HoveredStateLayerUnselectedOpacity: token.OpacityLevel2,
		HoveredIconSelectedColor:           materialTheme.Scheme.Primary.Color,
		HoveredIconUnselectedColor:         materialTheme.Scheme.Surface.OnColor,

		FocusedStateLayerSelectedColor:     materialTheme.Scheme.Primary.Color,
		FocusedStateLayerSelectedOpacity:   token.OpacityLevel3,
		FocusedStateLayerUnselectedColor:   materialTheme.Scheme.Surface.OnColor,
		FocusedStateLayerUnselectedOpacity: token.OpacityLevel3,
		FocusedIconSelectedColor:           materialTheme.Scheme.Primary.Color,
		FocusedIconUnselectedColor:         materialTheme.Scheme.Surface.OnColor,

		PressedStateLayerSelectedColor:     materialTheme.Scheme.Surface.OnColor,
		PressedStateLayerSelectedOpacity:   token.OpacityLevel3,
		PressedStateLayerUnselectedColor:   materialTheme.Scheme.Primary.Color,
		PressedStateLayerUnselectedOpacity: token.OpacityLevel3,
		PressedIconSelectedColor:           materialTheme.Scheme.Primary.Color,
		PressedIconUnselectedColor:         materialTheme.Scheme.Surface.OnColor,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
