// SPDX-License-Identifier: Unlicense OR MIT

package slider

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.slider.Slider"

type Theme struct {
	SliderStopIndicatorColor                  token.MatColor
	SliderStopIndicatorColorSelected          token.MatColor
	SliderStopIndicatorEndSpace               unit.Dp
	SliderStopIndicatorShape                  token.CornerShapes
	SliderStopIndicatorSize                   unit.Dp
	SliderStopIndicatorStartSpace             unit.Dp
	SliderActiveStopIndicatorContainerColor   token.MatColor
	SliderInactiveStopIndicatorContainerColor token.MatColor

	SliderActiveTrackColor            token.MatColor
	SliderActiveTrackHeight           unit.Dp
	SliderActiveTrackInnerCornerShape token.CornerShape
	SliderActiveTrackOuterCornerShape token.CornerShape
	SliderActiveTrackShape            token.CornerShapes
	SliderInactiveTrackColor          token.MatColor
	SliderInactiveTrackHeight         unit.Dp
	SliderInactiveTrackShape          token.CornerShapes

	SliderHandleColor               token.MatColor
	SliderHandleHeight              unit.Dp
	SliderHandleShape               token.CornerShapes
	SliderHandleWidth               unit.Dp
	SliderActiveHandleColor         token.MatColor
	SliderActiveHandleHeight        unit.Dp
	SliderActiveHandleLeadingSpace  unit.Dp
	SliderActiveHandlePadding       unit.Dp
	SliderActiveHandleShape         token.CornerShapes
	SliderActiveHandleTrailingSpace unit.Dp
	SliderActiveHandleWidth         unit.Dp
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		SliderStopIndicatorColor:         materialTheme.Scheme.PrimaryContainer.OnColor,
		SliderStopIndicatorColorSelected: materialTheme.Scheme.Primary.OnColor,
		SliderStopIndicatorShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		SliderStopIndicatorSize:                   unit.Dp(4),
		SliderStopIndicatorStartSpace:             unit.Dp(8),
		SliderStopIndicatorEndSpace:               unit.Dp(8),
		SliderActiveStopIndicatorContainerColor:   materialTheme.Scheme.Primary.Color,
		SliderInactiveStopIndicatorContainerColor: materialTheme.Scheme.Primary.Color,
		SliderActiveTrackColor:                    materialTheme.Scheme.Primary.Color,
		SliderActiveTrackHeight:                   unit.Dp(24),
		SliderActiveTrackInnerCornerShape: token.CornerShape{
			Kind: token.CornerKindRound,
			Size: unit.Dp(2),
		},
		SliderActiveTrackOuterCornerShape: token.CornerShape{
			AdaptToSize: true,
			Kind:        token.CornerKindRound,
		},
		SliderInactiveTrackColor:  materialTheme.Scheme.SecondaryContainer.Color,
		SliderInactiveTrackHeight: unit.Dp(24),
		SliderHandleColor:         materialTheme.Scheme.Primary.Color,
		SliderHandleHeight:        unit.Dp(44),
		SliderHandleShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		SliderHandleWidth:              unit.Dp(4),
		SliderActiveHandleColor:        materialTheme.Scheme.Primary.Color,
		SliderActiveHandleHeight:       unit.Dp(44),
		SliderActiveHandleLeadingSpace: unit.Dp(4),
		SliderActiveHandlePadding:      unit.Dp(4),
		SliderActiveHandleShape: token.UniformCornerShapes(token.CornerShape{
			Kind:        token.CornerKindRound,
			AdaptToSize: true,
		}),
		SliderActiveHandleTrailingSpace: unit.Dp(4),
		SliderActiveHandleWidth:         unit.Dp(2),
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
