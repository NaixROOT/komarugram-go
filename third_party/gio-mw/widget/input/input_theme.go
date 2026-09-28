// SPDX-License-Identifier: Unlicense OR MIT

package input

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/unit"
)

const Namespace = "gio-mw.widget.input.Input"

const (
	paddingBottom            = unit.Dp(8)
	paddingStart             = unit.Dp(16)
	paddingLeadingIconStart  = unit.Dp(12)
	paddingTrailingIconEnd   = unit.Dp(12)
	paddingEnd               = unit.Dp(16)
	paddingTop               = unit.Dp(8)
	paddingSupportingTextTop = unit.Dp(4)
	widthMin                 = unit.Dp(60)
)

type Theme struct {
	EnabledContainerColor  token.MatColor
	EnabledContainerHeight unit.Dp
	EnabledContainerShape  token.CornerShapes
	EnabledLabelTextColor  token.MatColor
	//EnabledLabelTextType          token.TypeInfo
	//EnabledLabelTextPopulatedType token.TypeInfo
	EnabledLeadingIconColor      token.MatColor
	EnabledLeadingIconSize       unit.Dp
	EnabledTrailingIconColor     token.MatColor
	EnabledTrailingIconSize      unit.Dp
	EnabledActiveIndicatorColor  token.MatColor
	EnabledActiveIndicatorHeight unit.Dp
	EnabledSupportingTextColor   token.MatColor
	//EnabledSupportingTextType    token.TypeInfo
	EnabledInputTextColor token.MatColor
	//EnabledInputTextType         token.TypeInfo

	DisabledContainerColor         token.MatColor
	DisabledContainerOpacity       token.OpacityLevel
	DisabledLabelTextColor         token.MatColor
	DisabledLabelTextOpacity       token.OpacityLevel
	DisabledLeadingIconColor       token.MatColor
	DisabledLeadingIconOpacity     token.OpacityLevel
	DisabledTrailingIconColor      token.MatColor
	DisabledTrailingIconOpacity    token.OpacityLevel
	DisabledSupportingTextColor    token.MatColor
	DisabledSupportingTextOpacity  token.OpacityLevel
	DisabledInputTextColor         token.MatColor
	DisabledInputTextOpacity       token.OpacityLevel
	DisabledActiveIndicatorColor   token.MatColor
	DisabledActiveIndicatorOpacity token.OpacityLevel
	DisabledActiveIndicatorHeight  unit.Dp

	HoveredStateLayerColor       token.MatColor
	HoveredStateLayerOpacity     token.OpacityLevel
	HoveredLabelTextColor        token.MatColor
	HoveredLeadingIconColor      token.MatColor
	HoveredTrailingIconColor     token.MatColor
	HoveredSupportingTextColor   token.MatColor
	HoveredInputTextColor        token.MatColor
	HoveredActiveIndicatorColor  token.MatColor
	HoveredActiveIndicatorHeight unit.Dp

	FocusedLabelTextColor        token.MatColor
	FocusedLeadingIconColor      token.MatColor
	FocusedTrailingIconColor     token.MatColor
	FocusedSupportingTextColor   token.MatColor
	FocusedInputTextColor        token.MatColor
	FocusedActiveIndicatorColor  token.MatColor
	FocusedActiveIndicatorHeight unit.Dp

	ErrorLabelTextColor       token.MatColor
	ErrorLeadingIconColor     token.MatColor
	ErrorTrailingIconColor    token.MatColor
	ErrorSupportingTextColor  token.MatColor
	ErrorInputTextColor       token.MatColor
	ErrorActiveIndicatorColor token.MatColor

	ErrorFocusedLabelTextColor       token.MatColor
	ErrorFocusedLeadingIconColor     token.MatColor
	ErrorFocusedTrailingIconColor    token.MatColor
	ErrorFocusedSupportingTextColor  token.MatColor
	ErrorFocusedInputTextColor       token.MatColor
	ErrorFocusedActiveIndicatorColor token.MatColor

	ErrorHoveredStateLayerColor      token.MatColor
	ErrorHoveredStateLayerOpacity    token.OpacityLevel
	ErrorHoveredLabelTextColor       token.MatColor
	ErrorHoveredLeadingIconColor     token.MatColor
	ErrorHoveredTrailingIconColor    token.MatColor
	ErrorHoveredSupportingTextColor  token.MatColor
	ErrorHoveredInputTextColor       token.MatColor
	ErrorHoveredActiveIndicatorColor token.MatColor
}

func BuildTheme(gtx layout.Context) *Theme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	wTheme := wdk.GetWidgetTheme[Theme](materialTheme, Namespace)
	if wTheme != nil {
		return wTheme
	}
	widgetTheme := &Theme{
		EnabledContainerColor:  materialTheme.Scheme.SurfaceContainerHighest,
		EnabledContainerHeight: unit.Dp(56),
		EnabledContainerShape: token.CornerShapes{
			TopStart: token.CornerShape{
				Kind: token.CornerKindRound,
				Size: unit.Dp(4),
			},
			TopEnd: token.CornerShape{
				Kind: token.CornerKindRound,
				Size: unit.Dp(4),
			},
		},
		EnabledLabelTextColor: materialTheme.Scheme.SurfaceVariant.OnColor,
		//EnabledLabelTextType:          materialTheme.Typescale.BodyLarge,
		//EnabledLabelTextPopulatedType: materialTheme.Typescale.BodySmall,
		EnabledLeadingIconColor:      materialTheme.Scheme.SurfaceVariant.OnColor,
		EnabledLeadingIconSize:       unit.Dp(24),
		EnabledTrailingIconColor:     materialTheme.Scheme.SurfaceVariant.OnColor,
		EnabledTrailingIconSize:      unit.Dp(24),
		EnabledActiveIndicatorColor:  materialTheme.Scheme.SurfaceVariant.OnColor,
		EnabledActiveIndicatorHeight: unit.Dp(1),
		EnabledSupportingTextColor:   materialTheme.Scheme.SurfaceVariant.OnColor,
		//EnabledSupportingTextType:    materialTheme.Typescale.BodySmall,
		EnabledInputTextColor: materialTheme.Scheme.Surface.OnColor,
		//EnabledInputTextType:  materialTheme.Typescale.BodyLarge,

		DisabledContainerColor:         materialTheme.Scheme.Surface.OnColor,
		DisabledContainerOpacity:       token.OpacityLevel2,
		DisabledLabelTextColor:         materialTheme.Scheme.Surface.OnColor,
		DisabledLabelTextOpacity:       token.OpacityLevel9,
		DisabledLeadingIconColor:       materialTheme.Scheme.Surface.OnColor,
		DisabledLeadingIconOpacity:     token.OpacityLevel9,
		DisabledTrailingIconColor:      materialTheme.Scheme.Surface.OnColor,
		DisabledTrailingIconOpacity:    token.OpacityLevel9,
		DisabledSupportingTextColor:    materialTheme.Scheme.Surface.OnColor,
		DisabledSupportingTextOpacity:  token.OpacityLevel9,
		DisabledInputTextColor:         materialTheme.Scheme.Surface.OnColor,
		DisabledInputTextOpacity:       token.OpacityLevel9,
		DisabledActiveIndicatorColor:   materialTheme.Scheme.SurfaceVariant.OnColor,
		DisabledActiveIndicatorOpacity: token.OpacityLevel9,
		DisabledActiveIndicatorHeight:  unit.Dp(1),

		HoveredStateLayerColor:       materialTheme.Scheme.Surface.OnColor,
		HoveredStateLayerOpacity:     token.OpacityLevel2,
		HoveredLabelTextColor:        materialTheme.Scheme.SurfaceVariant.OnColor,
		HoveredLeadingIconColor:      materialTheme.Scheme.SurfaceVariant.OnColor,
		HoveredTrailingIconColor:     materialTheme.Scheme.SurfaceVariant.OnColor,
		HoveredSupportingTextColor:   materialTheme.Scheme.SurfaceVariant.OnColor,
		HoveredInputTextColor:        materialTheme.Scheme.Surface.OnColor,
		HoveredActiveIndicatorColor:  materialTheme.Scheme.Surface.OnColor,
		HoveredActiveIndicatorHeight: unit.Dp(1),

		FocusedLabelTextColor:        materialTheme.Scheme.Primary.Color,
		FocusedLeadingIconColor:      materialTheme.Scheme.SurfaceVariant.OnColor,
		FocusedTrailingIconColor:     materialTheme.Scheme.SurfaceVariant.OnColor,
		FocusedSupportingTextColor:   materialTheme.Scheme.SurfaceVariant.OnColor,
		FocusedInputTextColor:        materialTheme.Scheme.Surface.OnColor,
		FocusedActiveIndicatorColor:  materialTheme.Scheme.Primary.Color,
		FocusedActiveIndicatorHeight: unit.Dp(2),

		ErrorLabelTextColor:       materialTheme.Scheme.Error.Color,
		ErrorLeadingIconColor:     materialTheme.Scheme.SurfaceVariant.OnColor,
		ErrorTrailingIconColor:    materialTheme.Scheme.Error.Color,
		ErrorSupportingTextColor:  materialTheme.Scheme.Error.Color,
		ErrorInputTextColor:       materialTheme.Scheme.Surface.OnColor,
		ErrorActiveIndicatorColor: materialTheme.Scheme.Error.Color,

		ErrorFocusedLabelTextColor:       materialTheme.Scheme.Error.Color,
		ErrorFocusedLeadingIconColor:     materialTheme.Scheme.SurfaceVariant.OnColor,
		ErrorFocusedTrailingIconColor:    materialTheme.Scheme.Error.Color,
		ErrorFocusedSupportingTextColor:  materialTheme.Scheme.Error.Color,
		ErrorFocusedInputTextColor:       materialTheme.Scheme.Surface.OnColor,
		ErrorFocusedActiveIndicatorColor: materialTheme.Scheme.Error.Color,

		ErrorHoveredStateLayerColor:      materialTheme.Scheme.Surface.OnColor,
		ErrorHoveredStateLayerOpacity:    token.OpacityLevel2,
		ErrorHoveredLabelTextColor:       materialTheme.Scheme.ErrorContainer.OnColor,
		ErrorHoveredLeadingIconColor:     materialTheme.Scheme.SurfaceVariant.OnColor,
		ErrorHoveredTrailingIconColor:    materialTheme.Scheme.ErrorContainer.OnColor,
		ErrorHoveredSupportingTextColor:  materialTheme.Scheme.Error.Color,
		ErrorHoveredInputTextColor:       materialTheme.Scheme.Surface.OnColor,
		ErrorHoveredActiveIndicatorColor: materialTheme.Scheme.ErrorContainer.OnColor,
	}
	wdk.InitWidgetThemeInTheme(materialTheme, Namespace, widgetTheme)
	return widgetTheme
}
