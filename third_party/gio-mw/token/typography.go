// SPDX-License-Identifier: Unlicense OR MIT

package token

// @see https://m3.material.io/styles/typography/type-scale-tokens

import (
	"gioui.org/font"
	"gioui.org/unit"
)

type TypeInfo struct {
	Name       string
	Font       font.Typeface
	LineHeight unit.Sp
	Size       unit.Sp
	Tracking   unit.Sp
	Weight     font.Weight
}

func (i TypeInfo) AsRegularFont() font.Font {
	return font.Font{
		Typeface: i.Font,
		Style:    font.Regular,
	}
}

type Typeface uint8

const (
	TypefaceDefault Typeface = iota
	TypefaceDisplay
	TypefaceHeadline
	TypefaceTitle
	TypefaceBody
	TypefaceLabel

	TypefaceBrand
	TypefacePreformatted

	// TypefaceCustomStart defines the starting index for custom Typeface constants, separating them from predefined ones.
	TypefaceCustomStart = 16
)

type TypefaceArray [32]font.Typeface

type Typestyle uint8

const (
	TypestyleDefault Typestyle = iota

	TypestyleLabelSmall
	TypestyleLabelMedium
	TypestyleLabelLarge
	TypestyleBodySmall
	TypestyleBodyMedium
	TypestyleBodyLarge
	TypestyleTitleSmall
	TypestyleTitleMedium
	TypestyleTitleLarge
	TypestyleHeadlineSmall
	TypestyleHeadlineMedium
	TypestyleHeadlineLarge
	TypestyleDisplaySmall
	TypestyleDisplayMedium
	TypestyleDisplayLarge

	TypestyleLabelSmallEmphasized
	TypestyleLabelMediumEmphasized
	TypestyleLabelLargeEmphasized
	TypestyleBodySmallEmphasized
	TypestyleBodyMediumEmphasized
	TypestyleBodyLargeEmphasized
	TypestyleTitleSmallEmphasized
	TypestyleTitleMediumEmphasized
	TypestyleTitleLargeEmphasized
	TypestyleHeadlineSmallEmphasized
	TypestyleHeadlineMediumEmphasized
	TypestyleHeadlineLargeEmphasized
	TypestyleDisplaySmallEmphasized
	TypestyleDisplayMediumEmphasized
	TypestyleDisplayLargeEmphasized

	TypestylePreformatted

	// TypestyleCustomStart defines the starting value for custom style constants, separating them from predefined styles.
	TypestyleCustomStart = 64
)

type TypescaleArray [128]TypeInfo
