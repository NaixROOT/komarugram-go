// SPDX-License-Identifier: Unlicense OR MIT

package button

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.button"
const ElevatedNamespace = "gio-mw.widget.button.Elevated"
const FilledNamespace = "gio-mw.widget.button.Filled"
const FilledTonalNamespace = "gio-mw.widget.button.FilledTonal"
const OutlinedNamespace = "gio-mw.widget.button.Outlined"
const TextNamespace = "gio-mw.widget.button.Text"

var (
	defaultMinWidth              = unit.Dp(48)
	paddingStart                 = unit.Dp(24)
	paddingStartWithIcon         = unit.Dp(16)
	paddingEnd                   = unit.Dp(24)
	paddingEndOfIcon             = unit.Dp(8)
	kindTextPaddingStart         = unit.Dp(12)
	kindTextPaddingStartWithIcon = unit.Dp(16)
	kindTextPaddingEnd           = unit.Dp(12)
	iconOnlyPadding              = unit.Dp(8)
)

type Theme struct {
	// TODO: Add button label font style properties.
	EnabledContainerShape       token.CornerShapes
	EnabledContainerHeight      unit.Dp
	EnabledContainerElevation   token.ElevationLevel
	EnabledContainerShadowColor token.MatColor
	EnabledContainerColor       token.MatColor
	EnabledOutlineColor         token.MatColor
	EnabledOutlineWidth         unit.Dp
	EnabledLabelColor           token.MatColor
	EnabledIconColor            token.MatColor
	EnabledIconSize             unit.Dp

	DisabledContainerColor     token.MatColor
	DisabledContainerElevation token.ElevationLevel
	DisabledContainerOpacity   token.OpacityLevel
	DisabledOutlineColor       token.MatColor
	DisabledOutlineOpacity     token.OpacityLevel
	DisabledLabelColor         token.MatColor
	DisabledLabelOpacity       token.OpacityLevel
	DisabledIconColor          token.MatColor
	DisabledIconOpacity        token.OpacityLevel

	HoveredContainerElevation token.ElevationLevel
	HoveredStateLayerColor    token.MatColor
	HoveredStateLayerOpacity  token.OpacityLevel
	HoveredOutlineColor       token.MatColor
	HoveredLabelColor         token.MatColor
	HoveredIconColor          token.MatColor

	FocusedFocusIndicatorColor     token.MatColor
	FocusedFocusIndicatorThickness unit.Dp
	FocusedFocusIndicatorOffset    unit.Dp
	FocusedContainerElevation      token.ElevationLevel
	FocusedStateLayerColor         token.MatColor
	FocusedStateLayerOpacity       token.OpacityLevel
	FocusedOutlineColor            token.MatColor
	FocusedLabelColor              token.MatColor
	FocusedIconColor               token.MatColor

	PressedContainerElevation token.ElevationLevel
	PressedContainerShape     token.CornerShapes
	PressedStateLayerColor    token.MatColor
	PressedStateLayerOpacity  token.OpacityLevel
	PressedOutlineColor       token.MatColor
	PressedLabelColor         token.MatColor
	PressedIconColor          token.MatColor

	DraggedContainerElevation token.ElevationLevel
	DraggedStateLayerColor    token.MatColor
	DraggedStateLayerOpacity  token.OpacityLevel
	DraggedLabelColor         token.MatColor
	DraggedIconColor          token.MatColor
}

// AlternativeColorScheme defines a custom color set for text button labels and icons when enabled.
// TODO: Add support for a limited set of alternative color schemes, for warning and error.
type AlternativeColorScheme struct {
	EnabledLabelColor token.MatColor
	EnabledIconColor  token.MatColor
}

func BuildElevatedTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, ElevatedNamespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		EnabledContainerHeight:      unit.Dp(40),
		EnabledContainerElevation:   token.ElevationLevel1,
		EnabledContainerShadowColor: materialTheme.Scheme.Shadow,
		EnabledContainerColor:       materialTheme.Scheme.SurfaceContainerLow,
		//EnabledOutlineColor:            materialTheme.Scheme.Primary.Color,
		//EnabledOutlineWidth:            0,
		EnabledLabelColor:          materialTheme.Scheme.Primary.Color,
		EnabledIconColor:           materialTheme.Scheme.Primary.Color,
		EnabledIconSize:            unit.Dp(18),
		DisabledContainerColor:     materialTheme.Scheme.Surface.OnColor,
		DisabledContainerElevation: token.ElevationLevel0,
		DisabledContainerOpacity:   token.OpacityLevel4,
		//DisabledOutlineColor:           materialTheme.Scheme.Primary.Color,
		//DisabledOutlineOpacity:         0,
		DisabledLabelColor:        materialTheme.Scheme.Surface.OnColor,
		DisabledLabelOpacity:      token.OpacityLevel9,
		DisabledIconColor:         materialTheme.Scheme.Surface.OnColor,
		DisabledIconOpacity:       token.OpacityLevel9,
		HoveredContainerElevation: token.ElevationLevel2,
		HoveredStateLayerColor:    materialTheme.Scheme.Primary.Color,
		HoveredStateLayerOpacity:  token.OpacityLevel4,
		//HoveredOutlineColor:            materialTheme.Scheme.Primary.Color,
		HoveredLabelColor:              materialTheme.Scheme.Primary.Color,
		HoveredIconColor:               materialTheme.Scheme.Primary.Color,
		FocusedFocusIndicatorColor:     materialTheme.Scheme.Secondary.Color,
		FocusedFocusIndicatorThickness: unit.Dp(3),
		FocusedFocusIndicatorOffset:    unit.Dp(2),
		FocusedContainerElevation:      token.ElevationLevel1,
		FocusedStateLayerColor:         materialTheme.Scheme.Primary.Color,
		FocusedStateLayerOpacity:       token.OpacityLevel4,
		//FocusedOutlineColor:            materialTheme.Scheme.Primary.Color,
		FocusedLabelColor: materialTheme.Scheme.Primary.Color,
		FocusedIconColor:  materialTheme.Scheme.Primary.Color,
		PressedContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(12),
		}),
		PressedContainerElevation: token.ElevationLevel1,
		PressedStateLayerColor:    materialTheme.Scheme.Primary.Color,
		PressedStateLayerOpacity:  token.OpacityLevel4,
		//PressedOutlineColor:            materialTheme.Scheme.Primary.Color,
		PressedLabelColor: materialTheme.Scheme.Primary.Color,
		PressedIconColor:  materialTheme.Scheme.Primary.Color,
		//DraggedContainerElevation: elevation.ElevationLevel1,
		//DraggedStateLayerColor:    materialTheme.Scheme.Primary.Color,
		//DraggedStateLayerOpacity:  0,
		//DraggedLabelColor:         materialTheme.Scheme.Primary.Color,
		//DraggedIconColor:          materialTheme.Scheme.Primary.Color,
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
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		EnabledContainerHeight:      unit.Dp(40),
		EnabledContainerElevation:   token.ElevationLevel0,
		EnabledContainerShadowColor: materialTheme.Scheme.Shadow,
		EnabledContainerColor:       materialTheme.Scheme.Primary.Color,
		//EnabledOutlineColor:            materialTheme.Scheme.Primary.Color,
		//EnabledOutlineWidth:            0,
		EnabledLabelColor:          materialTheme.Scheme.Primary.OnColor,
		EnabledIconColor:           materialTheme.Scheme.Primary.OnColor,
		EnabledIconSize:            unit.Dp(18),
		DisabledContainerColor:     materialTheme.Scheme.Surface.OnColor,
		DisabledContainerElevation: token.ElevationLevel0,
		DisabledContainerOpacity:   token.OpacityLevel4,
		//DisabledOutlineColor:           materialTheme.Scheme.Primary.Color,
		//DisabledOutlineOpacity:         0,
		DisabledLabelColor:        materialTheme.Scheme.Surface.OnColor,
		DisabledLabelOpacity:      token.OpacityLevel9,
		DisabledIconColor:         materialTheme.Scheme.Surface.OnColor,
		DisabledIconOpacity:       token.OpacityLevel9,
		HoveredContainerElevation: token.ElevationLevel1,
		HoveredStateLayerColor:    materialTheme.Scheme.Primary.OnColor,
		HoveredStateLayerOpacity:  token.OpacityLevel4,
		//HoveredOutlineColor:            materialTheme.Scheme.Primary.Color,
		HoveredLabelColor:              materialTheme.Scheme.Primary.OnColor,
		HoveredIconColor:               materialTheme.Scheme.Primary.OnColor,
		FocusedFocusIndicatorColor:     materialTheme.Scheme.Secondary.Color,
		FocusedFocusIndicatorThickness: unit.Dp(3),
		FocusedFocusIndicatorOffset:    unit.Dp(2),
		FocusedContainerElevation:      token.ElevationLevel0,
		FocusedStateLayerColor:         materialTheme.Scheme.Primary.OnColor,
		FocusedStateLayerOpacity:       token.OpacityLevel4,
		//FocusedOutlineColor:            materialTheme.Scheme.Primary.Color,
		FocusedLabelColor: materialTheme.Scheme.Primary.OnColor,
		FocusedIconColor:  materialTheme.Scheme.Primary.OnColor,
		PressedContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(12),
		}),
		PressedContainerElevation: token.ElevationLevel0,
		PressedStateLayerColor:    materialTheme.Scheme.Primary.OnColor,
		PressedStateLayerOpacity:  token.OpacityLevel4,
		//PressedOutlineColor:            materialTheme.Scheme.Primary.Color,
		PressedLabelColor:         materialTheme.Scheme.Primary.OnColor,
		PressedIconColor:          materialTheme.Scheme.Primary.OnColor,
		DraggedContainerElevation: token.ElevationLevel3,
		DraggedStateLayerColor:    materialTheme.Scheme.Primary.OnColor,
		DraggedStateLayerOpacity:  token.OpacityLevel5,
		DraggedLabelColor:         materialTheme.Scheme.Primary.OnColor,
		DraggedIconColor:          materialTheme.Scheme.Primary.OnColor,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, FilledNamespace, widgetTheme)
	return widgetTheme
}

func BuildFilledTonalTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, FilledTonalNamespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		EnabledContainerHeight:      unit.Dp(40),
		EnabledContainerElevation:   token.ElevationLevel0,
		EnabledContainerShadowColor: materialTheme.Scheme.Shadow,
		EnabledContainerColor:       materialTheme.Scheme.SecondaryContainer.Color,
		//EnabledOutlineColor:            materialTheme.Scheme.Primary.Color,
		//EnabledOutlineWidth:            0,
		EnabledLabelColor:          materialTheme.Scheme.SecondaryContainer.OnColor,
		EnabledIconColor:           materialTheme.Scheme.SecondaryContainer.OnColor,
		EnabledIconSize:            unit.Dp(18),
		DisabledContainerColor:     materialTheme.Scheme.Surface.OnColor,
		DisabledContainerElevation: token.ElevationLevel0,
		DisabledContainerOpacity:   token.OpacityLevel4,
		//DisabledOutlineColor:           materialTheme.Scheme.Primary.Color,
		//DisabledOutlineOpacity:         0,
		DisabledLabelColor:        materialTheme.Scheme.Surface.OnColor,
		DisabledLabelOpacity:      token.OpacityLevel9,
		DisabledIconColor:         materialTheme.Scheme.Surface.OnColor,
		DisabledIconOpacity:       token.OpacityLevel9,
		HoveredContainerElevation: token.ElevationLevel1,
		HoveredStateLayerColor:    materialTheme.Scheme.SecondaryContainer.OnColor,
		HoveredStateLayerOpacity:  token.OpacityLevel4,
		//HoveredOutlineColor:            materialTheme.Scheme.Primary.Color,
		HoveredLabelColor:              materialTheme.Scheme.SecondaryContainer.OnColor,
		HoveredIconColor:               materialTheme.Scheme.SecondaryContainer.OnColor,
		FocusedFocusIndicatorColor:     materialTheme.Scheme.Secondary.Color,
		FocusedFocusIndicatorThickness: unit.Dp(3),
		FocusedFocusIndicatorOffset:    unit.Dp(2),
		FocusedContainerElevation:      token.ElevationLevel0,
		FocusedStateLayerColor:         materialTheme.Scheme.SecondaryContainer.OnColor,
		FocusedStateLayerOpacity:       token.OpacityLevel4,
		//FocusedOutlineColor:            materialTheme.Scheme.Primary.Color,
		FocusedLabelColor: materialTheme.Scheme.SecondaryContainer.OnColor,
		FocusedIconColor:  materialTheme.Scheme.SecondaryContainer.OnColor,
		PressedContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(12),
		}),
		PressedContainerElevation: token.ElevationLevel0,
		PressedStateLayerColor:    materialTheme.Scheme.SecondaryContainer.OnColor,
		PressedStateLayerOpacity:  token.OpacityLevel4,
		//PressedOutlineColor:            materialTheme.Scheme.Primary.Color,
		PressedLabelColor: materialTheme.Scheme.SecondaryContainer.OnColor,
		PressedIconColor:  materialTheme.Scheme.SecondaryContainer.OnColor,
		//DraggedContainerElevation: elevation.ElevationLevel1,
		//DraggedStateLayerColor:    materialTheme.Scheme.Primary.Color,
		//DraggedStateLayerOpacity:  0,
		//DraggedLabelColor:         materialTheme.Scheme.Primary.Color,
		//DraggedIconColor:          materialTheme.Scheme.Primary.Color,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, FilledTonalNamespace, widgetTheme)
	return widgetTheme
}

func BuildOutlinedTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, OutlinedNamespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		EnabledContainerHeight:    unit.Dp(40),
		EnabledContainerElevation: token.ElevationLevel0,
		//EnabledContainerShadowColor: materialTheme.Scheme.Shadow,
		//EnabledContainerColor:       materialTheme.Scheme.SecondaryContainer.Color,
		EnabledOutlineColor: materialTheme.Scheme.Outline,
		EnabledOutlineWidth: unit.Dp(1),
		EnabledLabelColor:   materialTheme.Scheme.Primary.Color,
		EnabledIconColor:    materialTheme.Scheme.Primary.Color,
		EnabledIconSize:     unit.Dp(18),
		//DisabledContainerColor:     materialTheme.Scheme.Surface.OnColor,
		//DisabledContainerElevation: elevation.ElevationLevel0,
		//DisabledContainerOpacity:   token.OpacityLevel4,
		DisabledOutlineColor:   materialTheme.Scheme.Surface.OnColor,
		DisabledOutlineOpacity: token.OpacityLevel4,
		DisabledLabelColor:     materialTheme.Scheme.Surface.OnColor,
		DisabledLabelOpacity:   token.OpacityLevel9,
		DisabledIconColor:      materialTheme.Scheme.Surface.OnColor,
		DisabledIconOpacity:    token.OpacityLevel9,
		//HoveredContainerElevation: elevation.ElevationLevel1,
		HoveredStateLayerColor:         materialTheme.Scheme.Primary.Color,
		HoveredStateLayerOpacity:       token.OpacityLevel4,
		HoveredOutlineColor:            materialTheme.Scheme.Outline,
		HoveredLabelColor:              materialTheme.Scheme.Primary.Color,
		HoveredIconColor:               materialTheme.Scheme.Primary.Color,
		FocusedFocusIndicatorColor:     materialTheme.Scheme.Secondary.Color,
		FocusedFocusIndicatorThickness: unit.Dp(3),
		FocusedFocusIndicatorOffset:    unit.Dp(2),
		//FocusedContainerElevation:      elevation.ElevationLevel0,
		FocusedStateLayerColor:   materialTheme.Scheme.Primary.Color,
		FocusedStateLayerOpacity: token.OpacityLevel4,
		FocusedOutlineColor:      materialTheme.Scheme.Primary.Color,
		FocusedLabelColor:        materialTheme.Scheme.Primary.Color,
		FocusedIconColor:         materialTheme.Scheme.Primary.Color,
		PressedContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(12),
		}),
		//PressedContainerElevation: elevation.ElevationLevel0,
		PressedStateLayerColor:   materialTheme.Scheme.Primary.Color,
		PressedStateLayerOpacity: token.OpacityLevel4,
		PressedOutlineColor:      materialTheme.Scheme.Outline,
		PressedLabelColor:        materialTheme.Scheme.Primary.Color,
		PressedIconColor:         materialTheme.Scheme.Primary.Color,
		//DraggedContainerElevation: elevation.ElevationLevel1,
		//DraggedStateLayerColor:    materialTheme.Scheme.Primary.Color,
		//DraggedStateLayerOpacity:  0,
		//DraggedLabelColor:         materialTheme.Scheme.Primary.Color,
		//DraggedIconColor:          materialTheme.Scheme.Primary.Color,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, OutlinedNamespace, widgetTheme)
	return widgetTheme
}

func BuildTextTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, TextNamespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		EnabledContainerHeight:    unit.Dp(40),
		EnabledContainerElevation: token.ElevationLevel0,
		//EnabledContainerShadowColor: materialTheme.Scheme.Shadow,
		//EnabledContainerColor:       materialTheme.Scheme.SecondaryContainer.Color,
		//EnabledOutlineColor: materialTheme.Scheme.Outline,
		//EnabledOutlineWidth: unit.Dp(1),
		EnabledLabelColor: materialTheme.Scheme.Primary.Color,
		EnabledIconColor:  materialTheme.Scheme.Primary.Color,
		EnabledIconSize:   unit.Dp(18),
		//DisabledContainerColor:     materialTheme.Scheme.Surface.OnColor,
		//DisabledContainerElevation: elevation.ElevationLevel0,
		//DisabledContainerOpacity:   token.OpacityLevel4,
		//DisabledOutlineColor:       materialTheme.Scheme.Surface.OnColor,
		//DisabledOutlineOpacity:     token.OpacityLevel4,
		DisabledLabelColor:   materialTheme.Scheme.Surface.OnColor,
		DisabledLabelOpacity: token.OpacityLevel9,
		DisabledIconColor:    materialTheme.Scheme.Surface.OnColor,
		DisabledIconOpacity:  token.OpacityLevel9,
		//HoveredContainerElevation: elevation.ElevationLevel1,
		HoveredStateLayerColor:   materialTheme.Scheme.Primary.Color,
		HoveredStateLayerOpacity: token.OpacityLevel4,
		//HoveredOutlineColor:            materialTheme.Scheme.Outline,
		HoveredLabelColor:              materialTheme.Scheme.Primary.Color,
		HoveredIconColor:               materialTheme.Scheme.Primary.Color,
		FocusedFocusIndicatorColor:     materialTheme.Scheme.Secondary.Color,
		FocusedFocusIndicatorThickness: unit.Dp(3),
		FocusedFocusIndicatorOffset:    unit.Dp(2),
		//FocusedContainerElevation:      elevation.ElevationLevel0,
		FocusedStateLayerColor:   materialTheme.Scheme.Primary.Color,
		FocusedStateLayerOpacity: token.OpacityLevel4,
		//FocusedOutlineColor:      materialTheme.Scheme.Primary.Color,
		FocusedLabelColor: materialTheme.Scheme.Primary.Color,
		FocusedIconColor:  materialTheme.Scheme.Primary.Color,
		PressedContainerShape: token.UniformCornerShapes(token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(12),
		}),
		//PressedContainerElevation: elevation.ElevationLevel0,
		PressedStateLayerColor:   materialTheme.Scheme.Primary.Color,
		PressedStateLayerOpacity: token.OpacityLevel4,
		//PressedOutlineColor:      materialTheme.Scheme.Outline,
		PressedLabelColor: materialTheme.Scheme.Primary.Color,
		PressedIconColor:  materialTheme.Scheme.Primary.Color,
		//DraggedContainerElevation: elevation.ElevationLevel1,
		//DraggedStateLayerColor:    materialTheme.Scheme.Primary.Color,
		//DraggedStateLayerOpacity:  0,
		//DraggedLabelColor:         materialTheme.Scheme.Primary.Color,
		//DraggedIconColor:          materialTheme.Scheme.Primary.Color,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, TextNamespace, widgetTheme)
	return widgetTheme
}
