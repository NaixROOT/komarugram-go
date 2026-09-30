// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gio-mw/token"

	"gioui.org/layout"
	"gioui.org/op"
	"golang.org/x/image/font/gofont/gomono"

	"komarugram/internal/messenger/fonts"
	"komarugram/internal/messenger/preferences"
)

// TestFontOfTheSettingsReachesTheTheme checks that a font file saved in
// the settings is in the theme of the window's next frame, in the key of
// the measured messages, and gone from both once reset.
func TestFontOfTheSettingsReachesTheTheme(t *testing.T) {
	w := newSettingsWindow(t, "ru", preferences.ThemeLight, image.Pt(900, 700))
	defer w.close()
	t.Cleanup(func() { applyFonts(preferences.Fonts{}, nil) })
	gtx := layout.Context{Ops: new(op.Ops)}
	typeface := func() string { return string(w.a.Theme(gtx).Typescale[token.TypestyleBodyLarge].Font) }
	before, theme := typeface(), w.a.Theme(gtx)

	path := filepath.Join(t.TempDir(), "font.ttf")
	if err := os.WriteFile(path, gomono.TTF, 0o600); err != nil {
		t.Fatal(err)
	}
	w.a.settings.fontsView.setFiles(preferences.Fonts{Text: path})
	if got := typeface(); got != `"Go Mono", `+before {
		t.Errorf("typeface with a font picked: %q", got)
	}
	if w.a.Theme(gtx) == theme {
		t.Error("the window kept the theme made before the font was picked")
	}
	if fonts.Revision() == 1 {
		t.Error("messages measured with the font would be taken for the system font's")
	}
	if !w.frames("the appearance settings with a font", 2) {
		return
	}

	w.a.settings.fontsView.setFiles(preferences.Fonts{})
	if got := typeface(); got != before || strings.Contains(got, "Go Mono") {
		t.Errorf("typeface after a reset: %q, want %q", got, before)
	}
	if fonts.Revision() != 1 {
		t.Errorf("revision after a reset: %d", fonts.Revision())
	}
}
