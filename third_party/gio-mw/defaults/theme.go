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

	own := currentFonts()
	var shaperOptions []text.ShaperOption
	shaperOptions = append(shaperOptions, text.WithCollection(append(gofont.Collection(), own.Collection...)))
	if own.Emoji != "" {
		shaperOptions = append(shaperOptions, text.WithEmojiFamily(own.Emoji))
	}

	var typefaces token.TypefaceArray
	if own.EmojiImages != nil {
		shaperOptions = append(shaperOptions, text.WithEmojiImages(own.EmojiImages))
	}

	families := "Noto Sans, Liberation Sans, Noto Sans Arabic, Noto Sans Hebrew, Noto Color Emoji, sans-serif"
	if own.Emoji != "" {
		// The shaper draws emoji with the font of the program's own. Named
		// here, a font of this family would draw the digits of all text
		// on a system without the text fonts named before it: the families
		// named are tried before what stands in for sans-serif.
		families = "Noto Sans, Liberation Sans, Noto Sans Arabic, Noto Sans Hebrew, sans-serif"
	}
	typefaces[token.TypefaceDefault] = withFamilies(own.Default, families)
	typefaces[token.TypefacePreformatted] = withFamilies(own.Preformatted, "Noto Sans Mono, Liberation Mono, monospace")

	isDarkTheme := token.IsDarkColorSet(scheme.Background)

	return &token.Theme{
		Scheme:     scheme,
		TextShaper: text.NewShaper(shaperOptions...),
		Typescale:  NewTypescaleArray(typefaces, isDarkTheme),
		Widgets:    make(map[string]any),
	}
}
