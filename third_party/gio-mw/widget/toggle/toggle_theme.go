// SPDX-License-Identifier: Unlicense OR MIT

package toggle

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.toggle.Toggle"

const (
	iconStrokeWidth = unit.Dp(2)
)

type Theme struct {
	HandleShape            token.CornerShapes
	PressedHandleHeight    unit.Dp
	PressedHandleWidth     unit.Dp
	SelectedHandleColor    token.MatColor
	SelectedHandleHeight   unit.Dp
	SelectedHandleWidth    unit.Dp
	SelectedIconColor      token.MatColor
	SelectedIconSize       unit.Dp
	SelectedTrackColor     token.MatColor
	StateLayerColor        token.MatColor
	StateLayerSize         unit.Dp
	StateLayerShape        token.CornerShapes
	TrackHeight            unit.Dp
	TrackOutlineColor      token.MatColor
	TrackOutlineWidth      unit.Dp
	TrackShape             token.CornerShapes
	TrackWidth             unit.Dp
	UnselectedHandleColor  token.MatColor
	UnselectedHandleHeight unit.Dp
	UnselectedHandleWidth  unit.Dp
	UnselectedIconColor    token.MatColor
	UnselectedIconSize     unit.Dp
	UnselectedTrackColor   token.MatColor
	WithIconHandleHeight   unit.Dp
	WithIconHandleWidth    unit.Dp

	DisabledSelectedHandleColor         token.MatColor
	DisabledSelectedHandleOpacity       token.OpacityLevel
	DisabledSelectedIconColor           token.MatColor
	DisabledSelectedIconOpacity         token.OpacityLevel
	DisabledSelectedTrackColor          token.MatColor
	DisabledTrackOpacity                token.OpacityLevel
	DisabledUnselectedHandleColor       token.MatColor
	DisabledUnselectedHandleOpacity     token.OpacityLevel
	DisabledUnselectedIconColor         token.MatColor
	DisabledUnselectedIconOpacity       token.OpacityLevel
	DisabledUnselectedTrackColor        token.MatColor
	DisabledUnselectedTrackOutlineColor token.MatColor

	HoverSelectedHandleColor         token.MatColor
	HoverSelectedIconColor           token.MatColor
	HoverSelectedStateLayerColor     token.MatColor
	HoverSelectedStateLayerOpacity   token.OpacityLevel
	HoverSelectedTrackColor          token.MatColor
	HoverUnselectedHandleColor       token.MatColor
	HoverUnselectedIconColor         token.MatColor
	HoverUnselectedStateLayerColor   token.MatColor
	HoverUnselectedStateLayerOpacity token.OpacityLevel
	HoverUnselectedTrackColor        token.MatColor
	HoverUnselectedTrackOutlineColor token.MatColor

	FocusIndicatorColor              token.MatColor
	FocusIndicatorOffset             unit.Dp
	FocusIndicatorThickness          unit.Dp
	FocusSelectedHandleColor         token.MatColor
	FocusSelectedIconColor           token.MatColor
	FocusSelectedStateLayerColor     token.MatColor
	FocusSelectedStateLayerOpacity   token.OpacityLevel
	FocusSelectedTrackColor          token.MatColor
	FocusUnselectedHandleColor       token.MatColor
	FocusUnselectedIconColor         token.MatColor
	FocusUnselectedStateLayerColor   token.MatColor
	FocusUnselectedStateLayerOpacity token.OpacityLevel
	FocusUnselectedTrackColor        token.MatColor
	FocusUnselectedTrackOutlineColor token.MatColor

	PressedSelectedHandleColor         token.MatColor
	PressedSelectedIconColor           token.MatColor
	PressedSelectedStateLayerColor     token.MatColor
	PressedSelectedStateLayerOpacity   token.OpacityLevel
	PressedSelectedTrackColor          token.MatColor
	PressedUnselectedHandleColor       token.MatColor
	PressedUnselectedIconColor         token.MatColor
	PressedUnselectedStateLayerColor   token.MatColor
	PressedUnselectedStateLayerOpacity token.OpacityLevel
	PressedUnselectedTrackColor        token.MatColor
	PressedUnselectedTrackOutlineColor token.MatColor
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		HandleShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		PressedHandleHeight:  unit.Dp(28),
		PressedHandleWidth:   unit.Dp(28),
		SelectedHandleColor:  materialTheme.Scheme.Primary.OnColor,
		SelectedHandleHeight: unit.Dp(24),
		SelectedHandleWidth:  unit.Dp(24),
		SelectedIconColor:    materialTheme.Scheme.PrimaryContainer.OnColor,
		SelectedIconSize:     unit.Dp(16),
		SelectedTrackColor:   materialTheme.Scheme.Primary.Color,
		StateLayerColor:      materialTheme.Scheme.Tertiary.OnColor,
		StateLayerSize:       unit.Dp(40),
		StateLayerShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		TrackHeight:       unit.Dp(32),
		TrackOutlineColor: materialTheme.Scheme.Outline,
		TrackOutlineWidth: unit.Dp(2),
		TrackShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		TrackWidth:                          unit.Dp(52),
		UnselectedHandleColor:               materialTheme.Scheme.Outline,
		UnselectedHandleHeight:              unit.Dp(16),
		UnselectedHandleWidth:               unit.Dp(16),
		UnselectedIconColor:                 materialTheme.Scheme.SurfaceContainerHighest,
		UnselectedIconSize:                  unit.Dp(16),
		UnselectedTrackColor:                materialTheme.Scheme.SurfaceContainerHighest,
		WithIconHandleHeight:                unit.Dp(24),
		WithIconHandleWidth:                 unit.Dp(24),
		DisabledSelectedHandleColor:         materialTheme.Scheme.Surface.Color,
		DisabledSelectedHandleOpacity:       token.OpacityLevel10,
		DisabledSelectedIconColor:           materialTheme.Scheme.Surface.OnColor,
		DisabledSelectedIconOpacity:         token.OpacityLevel8,
		DisabledSelectedTrackColor:          materialTheme.Scheme.Surface.OnColor,
		DisabledTrackOpacity:                token.OpacityLevel4,
		DisabledUnselectedHandleColor:       materialTheme.Scheme.Surface.OnColor,
		DisabledUnselectedHandleOpacity:     token.OpacityLevel8,
		DisabledUnselectedIconColor:         materialTheme.Scheme.SurfaceContainerHighest,
		DisabledUnselectedIconOpacity:       token.OpacityLevel8,
		DisabledUnselectedTrackColor:        materialTheme.Scheme.SurfaceContainerHighest,
		DisabledUnselectedTrackOutlineColor: materialTheme.Scheme.Surface.OnColor,
		HoverSelectedHandleColor:            materialTheme.Scheme.PrimaryContainer.Color,
		HoverSelectedIconColor:              materialTheme.Scheme.PrimaryContainer.OnColor,
		HoverSelectedStateLayerColor:        materialTheme.Scheme.Primary.Color,
		HoverSelectedStateLayerOpacity:      token.OpacityLevel2,
		HoverSelectedTrackColor:             materialTheme.Scheme.Primary.Color,
		HoverUnselectedHandleColor:          materialTheme.Scheme.SurfaceVariant.OnColor,
		HoverUnselectedIconColor:            materialTheme.Scheme.SurfaceContainerHighest,
		HoverUnselectedStateLayerColor:      materialTheme.Scheme.Surface.OnColor,
		HoverUnselectedStateLayerOpacity:    token.OpacityLevel2,
		HoverUnselectedTrackColor:           materialTheme.Scheme.SurfaceContainerHighest,
		HoverUnselectedTrackOutlineColor:    materialTheme.Scheme.Outline,
		FocusIndicatorColor:                 materialTheme.Scheme.Secondary.Color,
		FocusIndicatorOffset:                unit.Dp(2),
		FocusIndicatorThickness:             unit.Dp(3),
		FocusSelectedHandleColor:            materialTheme.Scheme.PrimaryContainer.Color,
		FocusSelectedIconColor:              materialTheme.Scheme.PrimaryContainer.OnColor,
		FocusSelectedStateLayerColor:        materialTheme.Scheme.Primary.Color,
		FocusSelectedStateLayerOpacity:      token.OpacityLevel3,
		FocusSelectedTrackColor:             materialTheme.Scheme.Primary.Color,
		FocusUnselectedHandleColor:          materialTheme.Scheme.SurfaceVariant.OnColor,
		FocusUnselectedIconColor:            materialTheme.Scheme.SurfaceContainerHighest,
		FocusUnselectedStateLayerColor:      materialTheme.Scheme.Surface.OnColor,
		FocusUnselectedStateLayerOpacity:    token.OpacityLevel3,
		FocusUnselectedTrackColor:           materialTheme.Scheme.SurfaceContainerHighest,
		FocusUnselectedTrackOutlineColor:    materialTheme.Scheme.Outline,
		PressedSelectedHandleColor:          materialTheme.Scheme.PrimaryContainer.Color,
		PressedSelectedIconColor:            materialTheme.Scheme.PrimaryContainer.OnColor,
		PressedSelectedStateLayerColor:      materialTheme.Scheme.Primary.Color,
		PressedSelectedStateLayerOpacity:    token.OpacityLevel3,
		PressedSelectedTrackColor:           materialTheme.Scheme.Primary.Color,
		PressedUnselectedHandleColor:        materialTheme.Scheme.SurfaceVariant.OnColor,
		PressedUnselectedIconColor:          materialTheme.Scheme.SurfaceContainerHighest,
		PressedUnselectedStateLayerColor:    materialTheme.Scheme.Surface.OnColor,
		PressedUnselectedStateLayerOpacity:  token.OpacityLevel3,
		PressedUnselectedTrackColor:         materialTheme.Scheme.SurfaceContainerHighest,
		PressedUnselectedTrackOutlineColor:  materialTheme.Scheme.Outline,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
