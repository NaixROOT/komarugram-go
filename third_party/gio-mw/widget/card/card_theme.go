// SPDX-License-Identifier: Unlicense OR MIT

package card

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.card"
const ElevatedNamespace = "gio-mw.widget.card.Elevated"
const FilledNamespace = "gio-mw.widget.card.Filled"
const OutlinedNamespace = "gio-mw.widget.card.Outlined"

type Theme struct {
	EnabledContainerColor       token.MatColor
	EnabledContainerElevation   token.ElevationLevel
	EnabledContainerShape       token.CornerShapes
	EnabledContainerShadowColor token.MatColor
	EnabledOutlineShadowColor   token.MatColor
	EnabledOutlineColor         token.MatColor
	EnabledOutlineWidth         unit.Dp
	EnabledIconColor            token.MatColor
	EnabledIconSize             unit.Dp

	DisabledContainerColor     token.MatColor
	DisabledContainerElevation token.ElevationLevel
	DisabledContainerOpacity   token.OpacityLevel
	DisabledOutlineColor       token.MatColor
	DisabledOutlineOpacity     token.OpacityLevel

	HoveredContainerElevation token.ElevationLevel
	HoveredStateLayerColor    token.MatColor
	HoveredStateLayerOpacity  token.OpacityLevel
	HoveredOutlineColor       token.MatColor

	FocusedFocusIndicatorColor     token.MatColor
	FocusedFocusIndicatorThickness unit.Dp
	FocusedFocusIndicatorOffset    unit.Dp
	FocusedContainerElevation      token.ElevationLevel
	FocusedStateLayerColor         token.MatColor
	FocusedStateLayerOpacity       token.OpacityLevel
	FocusedOutlineColor            token.MatColor

	PressedContainerElevation token.ElevationLevel
	PressedStateLayerColor    token.MatColor
	PressedStateLayerOpacity  token.OpacityLevel
	PressedOutlineColor       token.MatColor

	DraggedContainerElevation token.ElevationLevel
	DraggedStateLayerColor    token.MatColor
	DraggedStateLayerOpacity  token.OpacityLevel
	DraggedOutlineColor       token.MatColor
}

func BuildElevatedTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, ElevatedNamespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerColor:     materialTheme.Scheme.SurfaceContainerLow,
		EnabledContainerElevation: token.ElevationLevel1,
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(12),
		}),
		EnabledContainerShadowColor: materialTheme.Scheme.Shadow,
		//EnabledOutlineShadowColor:      materialTheme.Scheme.Primary.Color,
		//EnabledOutlineColor:            materialTheme.Scheme.Primary.Color,
		//EnabledOutlineWidth:            0,
		EnabledIconColor:           materialTheme.Scheme.Primary.Color,
		EnabledIconSize:            unit.Dp(24),
		DisabledContainerColor:     materialTheme.Scheme.Surface.Color,
		DisabledContainerElevation: token.ElevationLevel1,
		DisabledContainerOpacity:   token.OpacityLevel9,
		//DisabledOutlineColor:           materialTheme.Scheme.Primary.Color,
		//DisabledOutlineOpacity:         0,
		HoveredContainerElevation: token.ElevationLevel2,
		HoveredStateLayerColor:    materialTheme.Scheme.Surface.OnColor,
		HoveredStateLayerOpacity:  token.OpacityLevel2,
		//HoveredOutlineColor:            materialTheme.Scheme.Primary.Color,
		FocusedFocusIndicatorColor:     materialTheme.Scheme.Secondary.Color,
		FocusedFocusIndicatorThickness: unit.Dp(3),
		FocusedFocusIndicatorOffset:    unit.Dp(2),
		FocusedContainerElevation:      token.ElevationLevel1,
		FocusedStateLayerColor:         materialTheme.Scheme.Surface.OnColor,
		FocusedStateLayerOpacity:       token.OpacityLevel3,
		//FocusedOutlineColor:            materialTheme.Scheme.Primary.Color,
		PressedContainerElevation: token.ElevationLevel1,
		PressedStateLayerColor:    materialTheme.Scheme.Surface.OnColor,
		PressedStateLayerOpacity:  token.OpacityLevel3,
		//PressedOutlineColor:            materialTheme.Scheme.Primary.Color,
		DraggedContainerElevation: token.ElevationLevel4,
		DraggedStateLayerColor:    materialTheme.Scheme.Surface.OnColor,
		DraggedStateLayerOpacity:  token.OpacityLevel5,
		//DraggedOutlineColor:            materialTheme.Scheme.Primary.Color,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, ElevatedNamespace, widgetTheme)
	return widgetTheme
}

func BuildFilledTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, FilledNamespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerColor:     materialTheme.Scheme.SurfaceContainerHighest,
		EnabledContainerElevation: token.ElevationLevel0,
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(12),
		}),
		EnabledContainerShadowColor: materialTheme.Scheme.Shadow,
		//EnabledOutlineShadowColor:      materialTheme.Scheme.Primary.Color,
		//EnabledOutlineColor:            materialTheme.Scheme.Primary.Color,
		//EnabledOutlineWidth:            0,
		EnabledIconColor:           materialTheme.Scheme.Primary.Color,
		EnabledIconSize:            unit.Dp(24),
		DisabledContainerColor:     materialTheme.Scheme.SurfaceVariant.Color,
		DisabledContainerElevation: token.ElevationLevel0,
		DisabledContainerOpacity:   token.OpacityLevel9,
		//DisabledOutlineColor:           materialTheme.Scheme.Primary.Color,
		//DisabledOutlineOpacity:         0,
		HoveredContainerElevation: token.ElevationLevel1,
		HoveredStateLayerColor:    materialTheme.Scheme.Surface.OnColor,
		HoveredStateLayerOpacity:  token.OpacityLevel2,
		//HoveredOutlineColor:            materialTheme.Scheme.Primary.Color,
		FocusedFocusIndicatorColor:     materialTheme.Scheme.Secondary.Color,
		FocusedFocusIndicatorThickness: unit.Dp(3),
		FocusedFocusIndicatorOffset:    unit.Dp(2),
		FocusedContainerElevation:      token.ElevationLevel0,
		FocusedStateLayerColor:         materialTheme.Scheme.Surface.OnColor,
		FocusedStateLayerOpacity:       token.OpacityLevel3,
		//FocusedOutlineColor:            materialTheme.Scheme.Primary.Color,
		PressedContainerElevation: token.ElevationLevel0,
		PressedStateLayerColor:    materialTheme.Scheme.Surface.OnColor,
		PressedStateLayerOpacity:  token.OpacityLevel3,
		//PressedOutlineColor:            materialTheme.Scheme.Primary.Color,
		DraggedContainerElevation: token.ElevationLevel3,
		DraggedStateLayerColor:    materialTheme.Scheme.Surface.OnColor,
		DraggedStateLayerOpacity:  token.OpacityLevel5,
		//DraggedOutlineColor:            materialTheme.Scheme.Primary.Color,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, FilledNamespace, widgetTheme)
	return widgetTheme
}

func BuildOutlinedTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, OutlinedNamespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerColor:     materialTheme.Scheme.Surface.Color,
		EnabledContainerElevation: token.ElevationLevel0,
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(12),
		}),
		//EnabledContainerShadowColor:    materialTheme.Scheme.Shadow,
		EnabledOutlineShadowColor: materialTheme.Scheme.Shadow,
		EnabledOutlineColor:       materialTheme.Scheme.OutlineVariant,
		EnabledOutlineWidth:       unit.Dp(1),
		EnabledIconColor:          materialTheme.Scheme.Primary.Color,
		EnabledIconSize:           unit.Dp(24),
		//DisabledContainerColor:         materialTheme.Scheme.SurfaceVariant.Color,
		DisabledContainerElevation: token.ElevationLevel0,
		//DisabledContainerOpacity: token.OpacityLevel9,
		DisabledOutlineColor:           materialTheme.Scheme.Outline,
		DisabledOutlineOpacity:         token.OpacityLevel4,
		HoveredContainerElevation:      token.ElevationLevel1,
		HoveredStateLayerColor:         materialTheme.Scheme.Surface.OnColor,
		HoveredStateLayerOpacity:       token.OpacityLevel2,
		HoveredOutlineColor:            materialTheme.Scheme.OutlineVariant,
		FocusedFocusIndicatorColor:     materialTheme.Scheme.Secondary.Color,
		FocusedFocusIndicatorThickness: unit.Dp(3),
		FocusedFocusIndicatorOffset:    unit.Dp(2),
		FocusedContainerElevation:      token.ElevationLevel0,
		FocusedStateLayerColor:         materialTheme.Scheme.Surface.OnColor,
		FocusedStateLayerOpacity:       token.OpacityLevel3,
		FocusedOutlineColor:            materialTheme.Scheme.Surface.OnColor,
		PressedContainerElevation:      token.ElevationLevel0,
		PressedStateLayerColor:         materialTheme.Scheme.Surface.OnColor,
		PressedStateLayerOpacity:       token.OpacityLevel3,
		PressedOutlineColor:            materialTheme.Scheme.OutlineVariant,
		DraggedContainerElevation:      token.ElevationLevel3,
		DraggedStateLayerColor:         materialTheme.Scheme.Surface.OnColor,
		DraggedStateLayerOpacity:       token.OpacityLevel5,
		DraggedOutlineColor:            materialTheme.Scheme.OutlineVariant,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, OutlinedNamespace, widgetTheme)
	return widgetTheme
}
