// SPDX-License-Identifier: Unlicense OR MIT

package token

// TODO: Add support for multiple static colors; Error is a static color.

// Scheme provides colors for MD3 themes.
// @see https://m3.material.io/styles/color/roles
// @see https://m3.material.io/styles/color/static/baseline
//
//go:generate go run ../cmd/genscheme ../defaults/schemes
type Scheme struct {
	Primary             MatColorSet
	PrimaryContainer    MatColorSet
	PrimaryFixed        MatColorSet
	PrimaryFixedVariant MatColorSet

	Secondary             MatColorSet
	SecondaryContainer    MatColorSet
	SecondaryFixed        MatColorSet
	SecondaryFixedVariant MatColorSet

	Tertiary             MatColorSet
	TertiaryContainer    MatColorSet
	TertiaryFixed        MatColorSet
	TertiaryFixedVariant MatColorSet

	Surface                 MatColorSet
	SurfaceVariant          MatColorSet
	SurfaceDim              MatColor
	SurfaceBright           MatColor
	SurfaceContainerLowest  MatColor
	SurfaceContainerLow     MatColor
	SurfaceContainer        MatColor
	SurfaceContainerHigh    MatColor
	SurfaceContainerHighest MatColor

	InverseSurface MatColorSet
	InversePrimary MatColor

	Background MatColorSet

	Outline        MatColor
	OutlineVariant MatColor

	Error          MatColorSet
	ErrorContainer MatColorSet

	Scrim  MatColor
	Shadow MatColor
}
