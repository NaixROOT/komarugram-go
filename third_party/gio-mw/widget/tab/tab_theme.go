// SPDX-License-Identifier: Unlicense OR MIT

package tab

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.tab.Tab"

type Theme struct {
	EnabledContainerColor        token.MatColor
	EnabledContainerElevation    token.ElevationLevel
	EnabledContainerHeight       unit.Dp
	EnabledContainerShape        token.CornerShapes
	EnabledLabelActiveColor      token.MatColor
	EnabledLabelInactiveColor    token.MatColor
	EnabledIconActiveColor       token.MatColor
	EnabledIconInactiveColor     token.MatColor
	EnabledIconSize              unit.Dp
	EnabledActiveIndicatorColor  token.MatColor
	EnabledActiveIndicatorHeight unit.Dp
	EnabledActiveIndicatorShape  token.CornerShapes

	HoveredLabelActiveColor          token.MatColor
	HoveredLabelInactiveColor        token.MatColor
	HoveredStateLayerActiveColor     token.MatColor
	HoveredStateLayerActiveOpacity   token.OpacityLevel
	HoveredStateLayerInactiveColor   token.MatColor
	HoveredStateLayerInactiveOpacity token.OpacityLevel
	HoveredIconActiveColor           token.MatColor
	HoveredIconInactiveColor         token.MatColor

	FocusedFocusIndicatorColor       token.MatColor
	FocusedFocusIndicatorThickness   unit.Dp
	FocusedFocusIndicatorOffset      unit.Dp
	FocusedLabelActiveColor          token.MatColor
	FocusedLabelInactiveColor        token.MatColor
	FocusedStateLayerActiveColor     token.MatColor
	FocusedStateLayerActiveOpacity   token.OpacityLevel
	FocusedStateLayerInactiveColor   token.MatColor
	FocusedStateLayerInactiveOpacity token.OpacityLevel
	FocusedIconActiveColor           token.MatColor
	FocusedIconInactiveColor         token.MatColor

	PressedLabelActiveColor          token.MatColor
	PressedLabelInactiveColor        token.MatColor
	PressedStateLayerActiveColor     token.MatColor
	PressedStateLayerActiveOpacity   token.OpacityLevel
	PressedStateLayerInactiveColor   token.MatColor
	PressedStateLayerInactiveOpacity token.OpacityLevel
	PressedIconActiveColor           token.MatColor
	PressedIconInactiveColor         token.MatColor
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerColor:     materialTheme.Scheme.Surface.Color,
		EnabledContainerElevation: token.ElevationLevel0,
		EnabledContainerHeight:    unit.Dp(48),
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindDefault,
			Size: unit.Dp(0),
		}),
		EnabledLabelActiveColor:      materialTheme.Scheme.Primary.Color,
		EnabledLabelInactiveColor:    materialTheme.Scheme.SurfaceVariant.OnColor,
		EnabledIconActiveColor:       materialTheme.Scheme.Primary.Color,
		EnabledIconInactiveColor:     materialTheme.Scheme.SurfaceVariant.OnColor,
		EnabledIconSize:              unit.Dp(24),
		EnabledActiveIndicatorColor:  materialTheme.Scheme.Primary.Color,
		EnabledActiveIndicatorHeight: unit.Dp(3),
		EnabledActiveIndicatorShape: token.CornerShapes{
			TopStart: token.CornerShape{
				Kind: token.CornerKindRound,
				Size: unit.Dp(3),
			},
			TopEnd: token.CornerShape{
				Kind: token.CornerKindRound,
				Size: unit.Dp(3),
			},
		},
		HoveredLabelActiveColor:          materialTheme.Scheme.Primary.Color,
		HoveredLabelInactiveColor:        materialTheme.Scheme.Surface.OnColor,
		HoveredStateLayerActiveColor:     materialTheme.Scheme.Primary.Color,
		HoveredStateLayerActiveOpacity:   token.OpacityLevel2,
		HoveredStateLayerInactiveColor:   materialTheme.Scheme.Surface.OnColor,
		HoveredStateLayerInactiveOpacity: token.OpacityLevel2,
		HoveredIconActiveColor:           materialTheme.Scheme.Primary.Color,
		HoveredIconInactiveColor:         materialTheme.Scheme.Surface.OnColor,
		FocusedFocusIndicatorColor:       materialTheme.Scheme.Secondary.Color,
		FocusedFocusIndicatorThickness:   unit.Dp(3),
		FocusedFocusIndicatorOffset:      unit.Dp(-3),
		FocusedLabelActiveColor:          materialTheme.Scheme.Primary.Color,
		FocusedLabelInactiveColor:        materialTheme.Scheme.Surface.OnColor,
		FocusedStateLayerActiveColor:     materialTheme.Scheme.Primary.Color,
		FocusedStateLayerActiveOpacity:   token.OpacityLevel3,
		FocusedStateLayerInactiveColor:   materialTheme.Scheme.Surface.OnColor,
		FocusedStateLayerInactiveOpacity: token.OpacityLevel3,
		FocusedIconActiveColor:           materialTheme.Scheme.Primary.Color,
		FocusedIconInactiveColor:         materialTheme.Scheme.Surface.OnColor,
		PressedLabelActiveColor:          materialTheme.Scheme.Primary.Color,
		PressedLabelInactiveColor:        materialTheme.Scheme.Surface.OnColor,
		PressedStateLayerActiveColor:     materialTheme.Scheme.Primary.Color,
		PressedStateLayerActiveOpacity:   token.OpacityLevel3,
		PressedStateLayerInactiveColor:   materialTheme.Scheme.Primary.Color,
		PressedStateLayerInactiveOpacity: token.OpacityLevel3,
		PressedIconActiveColor:           materialTheme.Scheme.Primary.Color,
		PressedIconInactiveColor:         materialTheme.Scheme.Surface.OnColor,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
