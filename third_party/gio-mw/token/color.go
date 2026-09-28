// SPDX-License-Identifier: Unlicense OR MIT

package token

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"
)

type MatColor color.NRGBA

func NewTransparentMatColor() MatColor {
	return MatColor{
		R: 0,
		G: 0,
		B: 0,
		A: 0,
	}
}

func NewMatColorFromHexRGB(c uint32) MatColor {
	maxVal := uint32(1<<24 - 1)
	if c > maxVal {
		panic(fmt.Sprintf("invalid RGB HEX color value: %v", c))
	}
	return MatColor{
		R: uint8(c >> 16),
		G: uint8(c >> 8),
		B: uint8(c),
		A: 0xFF,
	}
}

func NewMatColorFromHexRGBA(c uint32) MatColor {
	maxVal := uint32(1<<32 - 1)
	if c > maxVal {
		panic(fmt.Sprintf("invalid RGBA HEX color value: %v", c))
	}
	return MatColor{
		R: uint8(c >> 24),
		G: uint8(c >> 16),
		B: uint8(c >> 8),
		A: uint8(c),
	}
}

func NewMatColorFromHexString(c string) MatColor {
	lower := strings.ToLower(c)
	if lower[0] != '#' || len(lower) != 7 {
		panic(fmt.Sprintf("invalid RGB hex string: %v", c))
	}
	values, _ := strconv.ParseUint(lower[1:], 16, 32)
	col := color.RGBA{
		R: uint8(values >> 16),
		G: uint8(values >> 8),
		B: uint8(values),
		A: 0xFF,
	}
	return MatColor(col)
}

func (c MatColor) AsHex() string {
	return fmt.Sprintf("#%06X", c)
}

func (c MatColor) AsNRGBA() color.NRGBA {
	return color.NRGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

// SetOpacity applies opacity to the color.
func (c MatColor) SetOpacity(opacity OpacityLevel) MatColor {
	c.A = AsUint8Value(float32(opacity))
	return c
}

// LerpColor linearly interpolates from a to b, t in [0, 1]. The colors are
// mixed with premultiplied alpha, so fading to or from a transparent color
// keeps the hue of the visible one.
func LerpColor(a, b MatColor, t float32) MatColor {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	aA, bA := float32(a.A)/0xFF, float32(b.A)/0xFF
	alpha := aA + (bA-aA)*t
	if alpha <= 0 {
		return MatColor{}
	}
	mix := func(ca, cb uint8) uint8 {
		pa, pb := float32(ca)*aA, float32(cb)*bA
		return uint8(min(max((pa+(pb-pa)*t)/alpha+0.5, 0), 0xFF))
	}
	return MatColor{
		R: mix(a.R, b.R),
		G: mix(a.G, b.G),
		B: mix(a.B, b.B),
		A: uint8(min(alpha*0xFF+0.5, 0xFF)),
	}
}

// AsUint8Value converts a 0..1 float to a 0..255 uint8.
func AsUint8Value(value float32) uint8 {
	value *= 0xFF
	if value >= 0xFF {
		return 0xFF
	} else if value <= 0 {
		return 0x00
	}
	return uint8(value)
}

func Brightness(c MatColor) float32 {
	// @see https://alienryderflex.com/hsp.html
	// brightness = sqrt(0.299*R^2 + 0.587*G^2 + 0.114*B^2 )
	rBrightness := 0.299 * math.Pow(float64(c.R), 2)
	gBrightness := 0.587 * math.Pow(float64(c.G), 2)
	bBrightness := 0.114 * math.Pow(float64(c.B), 2)
	return float32(math.Sqrt(rBrightness + gBrightness + bBrightness))
}

func IsDarkColorSet(colorSet MatColorSet) bool {
	return Brightness(colorSet.Color) < Brightness(colorSet.OnColor)
}

type MatColorSet struct {
	Color   MatColor
	OnColor MatColor
}
