// SPDX-License-Identifier: Unlicense OR MIT

package defaults

import (
	"gio-mw/token"

	"gioui.org/font"
	"gioui.org/unit"
)

const (
	NativeOperatingSystemTypefaces = "BlinkMacSystemFont, Segoe UI, Roboto, Oxygen-Sans, Ubuntu, Cantarell, Helvetica Neue, Noto Sans, Noto Color Emoji, Noto Emoji, system-ui, sans-serif"
)

func NewTypescaleArray(typefaces token.TypefaceArray, isDarkTheme bool) *token.TypescaleArray {
	emphasizedWeight := font.Bold
	if isDarkTheme {
		// Reduce emphasis on dark themes.
		emphasizedWeight = font.Medium
	}

	var typescaleArray token.TypescaleArray

	if typefaces[token.TypefaceDefault] == "" {
		typefaces[token.TypefaceDefault] = NativeOperatingSystemTypefaces
	}
	typescaleArray[token.TypestyleDefault] = token.TypeInfo{
		Name:       "Default",
		Font:       typefaces[token.TypefaceDefault],
		Size:       unit.Sp(16),
		Tracking:   unit.Sp(0.5),
		LineHeight: unit.Sp(24),
	}

	if typefaces[token.TypefaceDisplay] == "" {
		typefaces[token.TypefaceDisplay] = typefaces[token.TypefaceDefault]
	}
	typescaleArray[token.TypestyleDisplayLarge] = token.TypeInfo{
		Name:       "Display Large",
		Font:       typefaces[token.TypefaceDisplay],
		Size:       unit.Sp(57),
		Tracking:   unit.Sp(-0.25),
		LineHeight: unit.Sp(64),
	}
	typescaleArray[token.TypestyleDisplayMedium] = token.TypeInfo{
		Name:       "Display Medium",
		Font:       typefaces[token.TypefaceDisplay],
		Size:       unit.Sp(45),
		Tracking:   unit.Sp(0),
		LineHeight: unit.Sp(52),
	}
	typescaleArray[token.TypestyleDisplaySmall] = token.TypeInfo{
		Name:       "Display Small",
		Font:       typefaces[token.TypefaceDisplay],
		Size:       unit.Sp(36),
		Tracking:   unit.Sp(0),
		LineHeight: unit.Sp(44),
	}
	typescaleArray[token.TypestyleDisplaySmallEmphasized] = token.TypeInfo{
		Name:       "Display Small Emphasized",
		Font:       typefaces[token.TypefaceDisplay],
		Size:       unit.Sp(36),
		Tracking:   unit.Sp(0),
		LineHeight: unit.Sp(44),
		Weight:     emphasizedWeight,
	}
	typescaleArray[token.TypestyleDisplayMediumEmphasized] = token.TypeInfo{
		Name:       "Display Medium Emphasized",
		Font:       typefaces[token.TypefaceDisplay],
		Size:       unit.Sp(45),
		Tracking:   unit.Sp(0),
		LineHeight: unit.Sp(52),
		Weight:     emphasizedWeight,
	}
	typescaleArray[token.TypestyleDisplayLargeEmphasized] = token.TypeInfo{
		Name:       "Display Large Emphasized",
		Font:       typefaces[token.TypefaceDisplay],
		Size:       unit.Sp(57),
		Tracking:   unit.Sp(-0.25),
		LineHeight: unit.Sp(64),
		Weight:     emphasizedWeight,
	}

	if typefaces[token.TypefaceHeadline] == "" {
		typefaces[token.TypefaceHeadline] = typefaces[token.TypefaceDefault]
	}
	typescaleArray[token.TypestyleHeadlineLarge] = token.TypeInfo{
		Name:       "Headline Large",
		Font:       typefaces[token.TypefaceHeadline],
		Size:       unit.Sp(32),
		Tracking:   unit.Sp(0),
		LineHeight: unit.Sp(40),
	}
	typescaleArray[token.TypestyleHeadlineMedium] = token.TypeInfo{
		Name:       "Headline Medium",
		Font:       typefaces[token.TypefaceHeadline],
		Size:       unit.Sp(28),
		Tracking:   unit.Sp(0),
		LineHeight: unit.Sp(36),
	}
	typescaleArray[token.TypestyleHeadlineSmall] = token.TypeInfo{
		Name:       "Headline Small",
		Font:       typefaces[token.TypefaceHeadline],
		Size:       unit.Sp(24),
		Tracking:   unit.Sp(0),
		LineHeight: unit.Sp(32),
	}
	typescaleArray[token.TypestyleHeadlineSmallEmphasized] = token.TypeInfo{
		Name:       "Headline Small Emphasized",
		Font:       typefaces[token.TypefaceHeadline],
		Size:       unit.Sp(24),
		Tracking:   unit.Sp(0),
		LineHeight: unit.Sp(32),
		Weight:     emphasizedWeight,
	}
	typescaleArray[token.TypestyleHeadlineMediumEmphasized] = token.TypeInfo{
		Name:       "Headline Medium Emphasized",
		Font:       typefaces[token.TypefaceHeadline],
		Size:       unit.Sp(28),
		Tracking:   unit.Sp(0),
		LineHeight: unit.Sp(36),
		Weight:     emphasizedWeight,
	}
	typescaleArray[token.TypestyleHeadlineLargeEmphasized] = token.TypeInfo{
		Name:       "Headline Large Emphasized",
		Font:       typefaces[token.TypefaceHeadline],
		Size:       unit.Sp(32),
		Tracking:   unit.Sp(0),
		LineHeight: unit.Sp(40),
		Weight:     emphasizedWeight,
	}

	if typefaces[token.TypefaceTitle] == "" {
		typefaces[token.TypefaceTitle] = typefaces[token.TypefaceDefault]
	}
	typescaleArray[token.TypestyleTitleLarge] = token.TypeInfo{
		Name:       "Title Large",
		Font:       typefaces[token.TypefaceTitle],
		Size:       unit.Sp(22),
		Tracking:   unit.Sp(0),
		LineHeight: unit.Sp(28),
	}
	typescaleArray[token.TypestyleTitleMedium] = token.TypeInfo{
		Name:       "Title Medium",
		Font:       typefaces[token.TypefaceTitle],
		Size:       unit.Sp(16),
		Tracking:   unit.Sp(0.15),
		LineHeight: unit.Sp(24),
	}
	typescaleArray[token.TypestyleTitleSmall] = token.TypeInfo{
		Name:       "Title Small",
		Font:       typefaces[token.TypefaceTitle],
		Size:       unit.Sp(14),
		Tracking:   unit.Sp(0.1),
		LineHeight: unit.Sp(20),
	}
	typescaleArray[token.TypestyleTitleSmallEmphasized] = token.TypeInfo{
		Name:       "Title Small Emphasized",
		Font:       typefaces[token.TypefaceTitle],
		Size:       unit.Sp(14),
		Tracking:   unit.Sp(0.1),
		LineHeight: unit.Sp(20),
		Weight:     emphasizedWeight,
	}
	typescaleArray[token.TypestyleTitleMediumEmphasized] = token.TypeInfo{
		Name:       "Title Medium Emphasized",
		Font:       typefaces[token.TypefaceTitle],
		Size:       unit.Sp(16),
		Tracking:   unit.Sp(0.15),
		LineHeight: unit.Sp(24),
		Weight:     emphasizedWeight,
	}
	typescaleArray[token.TypestyleTitleLargeEmphasized] = token.TypeInfo{
		Name:       "Title Large Emphasized",
		Font:       typefaces[token.TypefaceTitle],
		Size:       unit.Sp(22),
		Tracking:   unit.Sp(0),
		LineHeight: unit.Sp(28),
		Weight:     emphasizedWeight,
	}

	if typefaces[token.TypefaceBody] == "" {
		typefaces[token.TypefaceBody] = typefaces[token.TypefaceDefault]
	}
	typescaleArray[token.TypestyleBodyLarge] = token.TypeInfo{
		Name:       "Body Large",
		Font:       typefaces[token.TypefaceBody],
		Size:       unit.Sp(16),
		Tracking:   unit.Sp(0.5),
		LineHeight: unit.Sp(24),
	}
	typescaleArray[token.TypestyleBodyMedium] = token.TypeInfo{
		Name:       "Body Medium",
		Font:       typefaces[token.TypefaceBody],
		Size:       unit.Sp(14),
		Tracking:   unit.Sp(0.25),
		LineHeight: unit.Sp(20),
	}
	typescaleArray[token.TypestyleBodySmall] = token.TypeInfo{
		Name:       "Body Small",
		Font:       typefaces[token.TypefaceBody],
		Size:       unit.Sp(12),
		Tracking:   unit.Sp(0.4),
		LineHeight: unit.Sp(16),
	}
	typescaleArray[token.TypestyleBodySmallEmphasized] = token.TypeInfo{
		Name:       "Body Small Emphasized",
		Font:       typefaces[token.TypefaceBody],
		Size:       unit.Sp(12),
		Tracking:   unit.Sp(0.4),
		LineHeight: unit.Sp(16),
		Weight:     emphasizedWeight,
	}
	typescaleArray[token.TypestyleBodyMediumEmphasized] = token.TypeInfo{
		Name:       "Body Medium Emphasized",
		Font:       typefaces[token.TypefaceBody],
		Size:       unit.Sp(14),
		Tracking:   unit.Sp(0.25),
		LineHeight: unit.Sp(20),
		Weight:     emphasizedWeight,
	}
	typescaleArray[token.TypestyleBodyLargeEmphasized] = token.TypeInfo{
		Name:       "Body Large Emphasized",
		Font:       typefaces[token.TypefaceBody],
		Size:       unit.Sp(16),
		Tracking:   unit.Sp(0.5),
		LineHeight: unit.Sp(24),
		Weight:     emphasizedWeight,
	}

	if typefaces[token.TypefaceLabel] == "" {
		typefaces[token.TypefaceLabel] = typefaces[token.TypefaceDefault]
	}
	typescaleArray[token.TypestyleLabelLarge] = token.TypeInfo{
		Name:       "Label Large",
		Font:       typefaces[token.TypefaceLabel],
		Size:       unit.Sp(14),
		Tracking:   unit.Sp(0.1),
		LineHeight: unit.Sp(20),
	}

	typescaleArray[token.TypestyleLabelMedium] = token.TypeInfo{
		Name:       "Label Medium",
		Font:       typefaces[token.TypefaceLabel],
		Size:       unit.Sp(12),
		Tracking:   unit.Sp(0.5),
		LineHeight: unit.Sp(16),
	}

	typescaleArray[token.TypestyleLabelSmall] = token.TypeInfo{
		Name:       "Label Small",
		Font:       typefaces[token.TypefaceLabel],
		Size:       unit.Sp(11),
		Tracking:   unit.Sp(0.5),
		LineHeight: unit.Sp(16),
	}
	typescaleArray[token.TypestyleLabelSmallEmphasized] = token.TypeInfo{
		Name:       "Label Small Emphasized",
		Font:       typefaces[token.TypefaceLabel],
		Size:       unit.Sp(11),
		Tracking:   unit.Sp(0.5),
		LineHeight: unit.Sp(16),
		Weight:     emphasizedWeight,
	}
	typescaleArray[token.TypestyleLabelMediumEmphasized] = token.TypeInfo{
		Name:       "Label Medium Emphasized",
		Font:       typefaces[token.TypefaceLabel],
		Size:       unit.Sp(12),
		Tracking:   unit.Sp(0.5),
		LineHeight: unit.Sp(16),
		Weight:     emphasizedWeight,
	}
	typescaleArray[token.TypestyleLabelLargeEmphasized] = token.TypeInfo{
		Name:       "Label Large Emphasized",
		Font:       typefaces[token.TypefaceLabel],
		Size:       unit.Sp(14),
		Tracking:   unit.Sp(0.1),
		LineHeight: unit.Sp(20),
		Weight:     emphasizedWeight,
	}

	typescaleArray[token.TypestylePreformatted] = token.TypeInfo{
		Name:       "Preformatted",
		Font:       typefaces[token.TypefacePreformatted],
		Size:       unit.Sp(14),
		Tracking:   unit.Sp(0.25),
		LineHeight: unit.Sp(20),
	}

	return &typescaleArray
}
