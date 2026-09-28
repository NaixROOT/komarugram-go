// SPDX-License-Identifier: Unlicense OR MIT

package defaults

import (
	"gio-mw/token"

	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/text"
)

// NewTheme constructs a Theme (and underlying text shaper).
// layout.Context is required to force initialization after the first frame.
func NewTheme(gtx layout.Context, scheme *token.Scheme) *token.Theme {
	if scheme == nil {
		panic("NewTheme: scheme is required")
	}

	var shaperOptions []text.ShaperOption
	shaperOptions = append(shaperOptions, text.WithCollection(gofont.Collection()))

	var typefaces token.TypefaceArray
	typefaces[token.TypefaceDefault] = "Noto Sans, Liberation Sans, Noto Sans Arabic, Noto Sans Hebrew, Noto Color Emoji, sans-serif"
	typefaces[token.TypefacePreformatted] = "Noto Sans Mono, Liberation Mono, monospace"

	isDarkTheme := token.IsDarkColorSet(scheme.Background)

	return &token.Theme{
		Scheme:     scheme,
		TextShaper: text.NewShaper(shaperOptions...),
		Typescale:  NewTypescaleArray(typefaces, isDarkTheme),
		Widgets:    make(map[string]any),
	}
}
