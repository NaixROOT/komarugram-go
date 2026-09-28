// SPDX-License-Identifier: Unlicense OR MIT

package badge

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.badge.Badge"

type Theme struct {
	BadgeColor               token.MatColor
	BadgeShape               token.CornerShapes
	BadgeSize                unit.Dp
	BadgeLargeColor          token.MatColor
	BadgeLargeShape          token.CornerShapes
	BadgeLargeSize           unit.Dp
	BadgeLargeLabelTextColor token.MatColor
	BadgeLargeLabelTypestyle token.Typestyle
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		BadgeColor: materialTheme.Scheme.Error.Color,
		BadgeShape: token.UniformCornerShapes(
			token.CornerShape{Kind: token.CornerKindRound, AdaptToSize: true},
		),
		BadgeSize:       unit.Dp(6),
		BadgeLargeColor: materialTheme.Scheme.Error.Color,
		BadgeLargeShape: token.UniformCornerShapes(
			token.CornerShape{Kind: token.CornerKindRound, AdaptToSize: true},
		),
		BadgeLargeSize:           unit.Dp(16),
		BadgeLargeLabelTextColor: materialTheme.Scheme.Error.OnColor,
		BadgeLargeLabelTypestyle: token.TypestyleLabelSmall,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
