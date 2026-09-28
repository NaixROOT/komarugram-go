// SPDX-License-Identifier: Unlicense OR MIT

package sheet

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

var (
	maxWidth = unit.Dp(640)
)

const Namespace = "gio-mw.widget.sheet.Sheet"

type Theme struct {
	DockedContainerColor             token.MatColor
	DockedModalContainerElevation    token.ElevationLevel
	DockedStandardContainerElevation token.ElevationLevel
	BottomDockedContainerShape       token.CornerShapes
	SideDockedContainerShape         token.CornerShapes
	DockedDragHandleColor            token.MatColor
	DockedDragHandleWidth            unit.Dp
	DockedDragHandleHeight           unit.Dp
	FocusIndicatorColor              token.MatColor
	FocusIndicatorThickness          unit.Dp
	FocusIndicatorOffset             unit.Dp
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		DockedContainerColor:             materialTheme.Scheme.SurfaceContainerLow,
		DockedModalContainerElevation:    token.ElevationLevel1,
		DockedStandardContainerElevation: token.ElevationLevel1,
		BottomDockedContainerShape: token.CornerShapes{
			TopStart: token.CornerShape{
				Kind: token.CornerKindRound,
				Size: unit.Dp(28),
			},
			TopEnd: token.CornerShape{
				Kind: token.CornerKindRound,
				Size: unit.Dp(28),
			},
		},
		SideDockedContainerShape: token.CornerShapes{
			TopStart: token.CornerShape{
				Kind: token.CornerKindRound,
				Size: unit.Dp(28),
			},
			BottomStart: token.CornerShape{
				Kind: token.CornerKindRound,
				Size: unit.Dp(28),
			},
		},
		DockedDragHandleColor:   materialTheme.Scheme.SurfaceVariant.OnColor,
		DockedDragHandleWidth:   unit.Dp(32),
		DockedDragHandleHeight:  unit.Dp(4),
		FocusIndicatorColor:     materialTheme.Scheme.Secondary.Color,
		FocusIndicatorThickness: unit.Dp(3),
		FocusIndicatorOffset:    unit.Dp(2),
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
