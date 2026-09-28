// SPDX-License-Identifier: Unlicense OR MIT

package rail

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.rail.Rail"

type Theme struct {
	CollapsedContainerColor           token.MatColor
	ExpandedContainerWidthMinimum     unit.Dp
	ExpandedContainerWidthMaximum     unit.Dp
	ExpandedModalContainerShape       token.CornerShapes
	ItemActiveIndicatorColor          token.MatColor
	ItemIconSize                      unit.Dp
	ItemActiveIndicatorShape          token.CornerShapes
	ItemActiveIndicatorLeadingSpace   unit.Dp
	ItemActiveIndicatorIconLabelSpace unit.Dp
	ItemActiveIndicatorTrailingSpace  unit.Dp
	ItemContainerHeight               unit.Dp
	ItemContainerShape                token.CornerShapes
	ItemContainerVerticalSpace        unit.Dp
	ItemHeaderSpaceMinimum            unit.Dp
	ItemInactiveLabelColor            token.MatColor
	ItemActiveLabelTextColor          token.MatColor
	// The state layer of an item uses its label color with these opacities.
	ItemHoveredStateLayerOpacity token.OpacityLevel
	ItemPressedStateLayerOpacity token.OpacityLevel
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		CollapsedContainerColor:       materialTheme.Scheme.Surface.Color,
		ExpandedContainerWidthMinimum: unit.Dp(220),
		ExpandedContainerWidthMaximum: unit.Dp(360),
		ExpandedModalContainerShape: token.CornerShapes{
			TopEnd:    token.CornerShape{Kind: token.CornerKindRound, Size: unit.Dp(16)},
			BottomEnd: token.CornerShape{Kind: token.CornerKindRound, Size: unit.Dp(16)},
		},
		ItemActiveIndicatorColor: materialTheme.Scheme.SecondaryContainer.Color,
		ItemIconSize:             unit.Dp(24),
		ItemActiveIndicatorShape: token.UniformCornerShapes(
			token.CornerShape{Kind: token.CornerKindRound, AdaptToSize: true},
		),
		ItemActiveIndicatorLeadingSpace:   unit.Dp(16),
		ItemActiveIndicatorIconLabelSpace: unit.Dp(8),
		ItemActiveIndicatorTrailingSpace:  unit.Dp(16),
		ItemContainerHeight:               unit.Dp(64),
		ItemContainerShape:                token.CornerShapes{},
		ItemContainerVerticalSpace:        unit.Dp(6),
		ItemHeaderSpaceMinimum:            unit.Dp(40),
		ItemActiveLabelTextColor:          materialTheme.Scheme.SecondaryContainer.OnColor,
		ItemInactiveLabelColor:            materialTheme.Scheme.SurfaceVariant.OnColor,
		ItemHoveredStateLayerOpacity:      token.OpacityLevel2,
		ItemPressedStateLayerOpacity:      token.OpacityLevel3,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
