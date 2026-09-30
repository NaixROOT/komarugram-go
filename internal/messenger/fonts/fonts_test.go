// SPDX-License-Identifier: Unlicense OR MIT

package fonts

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/token"

	"gioui.org/layout"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/goregular"
)

// write puts a font into a file of the test's.
func write(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// typefaces are the families a theme made now asks for, for text and for
// preformatted text.
func typefaces() (text, mono string) {
	theme := defaults.NewTheme(layout.Context{}, schemes.SchemeBaselineLight())
	return string(theme.Typescale[token.TypestyleBodyLarge].Font), string(theme.Typescale[token.TypestylePreformatted].Font)
}

// reset returns the themes to the system's fonts after the test.
func reset(t *testing.T) {
	t.Cleanup(func() { Apply(Files{}) })
}

func TestApplyPutsTheFilesFirst(t *testing.T) {
	reset(t)
	text0, mono0 := typefaces()
	version := defaults.FontsVersion()
	files := Files{Text: write(t, "text.ttf", goregular.TTF), Mono: write(t, "mono.ttf", gomono.TTF)}
	if err := Apply(files); err != nil {
		t.Fatal(err)
	}
	if defaults.FontsVersion() == version {
		t.Error("the version of the fonts did not change")
	}
	text, mono := typefaces()
	if want := `"Go", ` + text0; text != want {
		t.Errorf("text typeface %q, want %q", text, want)
	}
	if want := `"Go Mono", ` + mono0; mono != want {
		t.Errorf("preformatted typeface %q, want %q", mono, want)
	}
	if Revision() == 1 {
		t.Error("the revision is the system fonts'")
	}
	// The same files again change nothing: every window applies them.
	version, rev := defaults.FontsVersion(), Revision()
	if err := Apply(files); err != nil {
		t.Fatal(err)
	}
	if defaults.FontsVersion() != version || Revision() != rev {
		t.Error("applying the same files made new fonts")
	}
	if err := Apply(Files{}); err != nil {
		t.Fatal(err)
	}
	if text, mono := typefaces(); text != text0 || mono != mono0 || Revision() != 1 {
		t.Errorf("after a reset: %q, %q, revision %d", text, mono, Revision())
	}
}

func TestExtraFollowsText(t *testing.T) {
	reset(t)
	text0, _ := typefaces()
	if err := Apply(Files{Extra: write(t, "extra.ttf", gomono.TTF), Text: write(t, "text.ttf", goregular.TTF)}); err != nil {
		t.Fatal(err)
	}
	if text, _ := typefaces(); text != `"Go", "Go Mono", `+text0 {
		t.Errorf("text typeface %q", text)
	}
}

func TestBadFilesArePassedOver(t *testing.T) {
	reset(t)
	text0, _ := typefaces()
	files := Files{
		Text:  write(t, "text.ttf", []byte("not a font")),
		Extra: filepath.Join(t.TempDir(), "gone.ttf"),
		Mono:  write(t, "mono.ttf", gomono.TTF),
		// A text font has no emoji.
		Emoji: write(t, "emoji.ttf", goregular.TTF),
	}
	err := Apply(files)
	if err == nil {
		t.Fatal("no error for files that do not load")
	}
	if !errors.Is(err, ErrNoEmoji) || !errors.Is(err, os.ErrNotExist) {
		t.Errorf("errors: %v", err)
	}
	text, mono := typefaces()
	if text != text0 {
		t.Errorf("text typeface %q, want the system's", text)
	}
	if !strings.HasPrefix(mono, `"Go Mono", `) {
		t.Errorf("preformatted typeface %q, want the good file's first", mono)
	}
}

func TestCheck(t *testing.T) {
	path := write(t, "text.ttf", goregular.TTF)
	f, err := Check(Text, path)
	if err != nil || f.Family != "Go" {
		t.Errorf("a text font: %q, %v", f.Family, err)
	}
	if _, err := Check(Emoji, path); !errors.Is(err, ErrNoEmoji) {
		t.Errorf("a text font for emoji: %v", err)
	}
	large := filepath.Join(t.TempDir(), "large.ttf")
	file, err := os.Create(large)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(MaxSize + 1); err != nil {
		t.Fatal(err)
	}
	file.Close()
	if _, err := Check(Text, large); !errors.Is(err, ErrTooLarge) {
		t.Errorf("a file over the limit: %v", err)
	}
}

func TestEnvironmentOverTheSettings(t *testing.T) {
	reset(t)
	text0, _ := typefaces()
	t.Setenv("KOMARUGRAM_FONT", write(t, "env.ttf", gomono.TTF))
	if !FromEnv(Text) || FromEnv(Emoji) {
		t.Error("FromEnv does not tell the role the environment names")
	}
	if err := Apply(Files{Text: write(t, "text.ttf", goregular.TTF)}); err != nil {
		t.Fatal(err)
	}
	if text, _ := typefaces(); text != `"Go Mono", `+text0 {
		t.Errorf("text typeface %q, want the environment's font first", text)
	}
}

// TestEmojiFontStaysOutOfTheTypeface checks that with an emoji font of the
// program's own no emoji family is named for text: on a system without the
// text fonts the theme names, a font of that family would draw the digits
// of every text. The shaper draws emoji with the font instead.
func TestEmojiFontStaysOutOfTheTypeface(t *testing.T) {
	t.Cleanup(func() { defaults.SetFonts(defaults.Fonts{}) })
	text0, _ := typefaces()
	if !strings.Contains(text0, "Emoji") {
		t.Skip("the theme names no emoji family")
	}
	defaults.SetFonts(defaults.Fonts{Emoji: "Noto Color Emoji"})
	if text, _ := typefaces(); strings.Contains(text, "Emoji") {
		t.Errorf("text typeface with an emoji font: %q", text)
	}
}
