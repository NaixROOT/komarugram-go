// SPDX-License-Identifier: Unlicense OR MIT

package defaults

import (
	"strings"
	"sync/atomic"

	"gioui.org/font"
)

// Fonts are fonts of the program's own that every theme made after
// SetFonts uses before the system's.
type Fonts struct {
	// Collection is loaded into the theme's shaper beside the Go fonts.
	Collection []font.FontFace
	// Default and Preformatted are families of the collection tried, in
	// this order, before the theme's own for text and for preformatted
	// text.
	Default, Preformatted []string
	// Emoji is the family of the collection that draws emoji, "" for the
	// system's.
	Emoji string
}

type fontsState struct {
	fonts   Fonts
	version uint64
}

var fonts atomic.Pointer[fontsState]

// SetFonts changes the fonts of the themes made from now on. Themes made
// before keep theirs: FontsVersion tells their holders to make new ones.
func SetFonts(f Fonts) {
	for {
		old := fonts.Load()
		next := &fontsState{fonts: f, version: 1}
		if old != nil {
			next.version = old.version + 1
		}
		if fonts.CompareAndSwap(old, next) {
			return
		}
	}
}

// FontsVersion changes with every SetFonts.
func FontsVersion() uint64 {
	if s := fonts.Load(); s != nil {
		return s.version
	}
	return 0
}

func currentFonts() Fonts {
	if s := fonts.Load(); s != nil {
		return s.fonts
	}
	return Fonts{}
}

// withFamilies puts families before the typeface's own.
func withFamilies(families []string, typeface string) font.Typeface {
	var b strings.Builder
	for _, f := range families {
		// A comma or a quote in a name would end it.
		b.WriteString(`"` + strings.NewReplacer(`"`, "", `\`, "").Replace(f) + `", `)
	}
	return font.Typeface(b.String() + typeface)
}
