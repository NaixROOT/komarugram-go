// SPDX-License-Identifier: Unlicense OR MIT

// Package fonts loads the fonts the user points at in the settings, and
// gives them to the themes before the system's.
package fonts

import (
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"gio-mw/defaults"

	"gioui.org/font"
	"gioui.org/font/opentype"
	gofont "github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/font/opentype/tables"

	"komarugram/internal/messenger/preferences"
)

// Files are the font files of the settings.
type Files = preferences.Fonts

// Role is what a font of Files draws.
type Role int

const (
	Text Role = iota
	Extra
	Mono
	Emoji
)

// Roles are the roles in the order the settings show them.
var Roles = [...]Role{Text, Extra, Mono, Emoji}

// Of returns the file of a role.
func Of(f Files, r Role) string { return *of(&f, r) }

// With returns f with the file of a role changed.
func With(f Files, r Role, path string) Files {
	*of(&f, r) = path
	return f
}

func of(f *Files, r Role) *string {
	switch r {
	case Extra:
		return &f.Extra
	case Mono:
		return &f.Mono
	case Emoji:
		return &f.Emoji
	}
	return &f.Text
}

// The variables that name a font file for a role over the settings', for
// trying a font and for the render tests.
var envNames = map[Role]string{
	Text:  "KOMARUGRAM_FONT",
	Extra: "KOMARUGRAM_FONT_EXTRA",
	Mono:  "KOMARUGRAM_FONT_MONO",
	Emoji: "KOMARUGRAM_FONT_EMOJI",
}

// WithEnv returns f with the files the environment names in place of its
// own.
func WithEnv(f Files) Files {
	for _, r := range Roles {
		if path := os.Getenv(envNames[r]); path != "" {
			f = With(f, r, path)
		}
	}
	return f
}

// FromEnv reports whether the environment names the file of a role, which
// the settings then do not change.
func FromEnv(r Role) bool { return os.Getenv(envNames[r]) != "" }

// MaxSize is the largest font file loaded: a file is kept in memory whole,
// and the largest CJK fonts of one weight are under it.
const MaxSize = 64 << 20

// ErrNoEmoji is the error of a font picked for emoji that has none.
var ErrNoEmoji = errors.New("the font has no emoji")

// ErrColorFormat is the error of a font picked for emoji whose glyphs are
// COLR version 1 paints, with gradients, which the shaper does not draw.
var ErrColorFormat = errors.New("the font's color glyphs are COLR version 1")

// ErrTooLarge is the error of a font file over MaxSize.
var ErrTooLarge = errors.New("the font file is too large")

// Font is a font file, loaded.
type Font struct {
	// Family is the family of the file's first face, and Families those of
	// all its faces, a collection's in the file's order.
	Family   string
	Families []string
	Faces    []font.FontFace
}

type loaded struct {
	size    int64
	modTime time.Time
	font    Font
}

var (
	mu    sync.Mutex
	cache = map[string]loaded{}
	// applied is what the themes have now.
	applied    Files
	appliedSet bool
)

// revision tells the fonts applied from others, and lasts across runs.
var revision atomic.Uint32

// Revision is a number of the fonts the themes use now, the same for the
// same files in every run, and 1 for the system's fonts: measurements of
// text kept on disk are of one revision.
func Revision() uint32 {
	if r := revision.Load(); r != 0 {
		return r
	}
	return 1
}

// Load parses a font file, once for the file as it is on disk: the faces
// are shared by the shapers of all windows.
func Load(path string) (Font, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Font{}, err
	}
	if info.Size() > MaxSize {
		return Font{}, ErrTooLarge
	}
	mu.Lock()
	defer mu.Unlock()
	if l, ok := cache[path]; ok && l.size == info.Size() && l.modTime.Equal(info.ModTime()) {
		return l.font, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Font{}, err
	}
	faces, err := opentype.ParseCollection(data)
	if err != nil {
		return Font{}, err
	}
	if len(faces) == 0 {
		return Font{}, errors.New("no fonts in the file")
	}
	f := Font{Family: string(faces[0].Font.Typeface), Faces: faces}
	for _, face := range faces {
		if family := string(face.Font.Typeface); family != "" && !slices.Contains(f.Families, family) {
			f.Families = append(f.Families, family)
		}
	}
	if f.Family == "" {
		return Font{}, errors.New("the font has no family name")
	}
	cache[path] = loaded{size: info.Size(), modTime: info.ModTime(), font: f}
	return f, nil
}

// Check loads a font file for a role, and turns down one that cannot fill
// it.
func Check(role Role, path string) (Font, error) {
	f, err := Load(path)
	if err != nil {
		return Font{}, err
	}
	if role == Emoji {
		if err := checkEmoji(f); err != nil {
			mu.Lock()
			delete(cache, path)
			mu.Unlock()
			return Font{}, err
		}
	}
	return f, nil
}

// checkEmoji makes sure the first face of f, the one the shaper takes for
// emoji, has them in a form it draws: bitmaps, outlines, or the layers of
// COLR version 0.
func checkEmoji(f Font) error {
	parsed, ok := f.Faces[0].Face.(opentype.Face)
	if !ok {
		return ErrNoEmoji
	}
	face := parsed.Face()
	gid, ok := face.NominalGlyph('\U0001F600')
	if !ok {
		return ErrNoEmoji
	}
	if glyph, ok := face.GlyphData(gid).(gofont.GlyphColor); ok {
		if _, ok := glyph.Paint.(tables.PaintColrLayersResolved); !ok {
			return ErrColorFormat
		}
	}
	return nil
}

// Apply makes the themes created from now on use the fonts of files, with
// the environment's over them. A file that does not load is left out, and
// its error returned with the others'; the rest still apply. Windows see
// the change by defaults.FontsVersion.
func Apply(files Files) error {
	files = WithEnv(files)
	mu.Lock()
	same := appliedSet && applied == files
	applied, appliedSet = files, true
	mu.Unlock()
	if same {
		return nil
	}
	var (
		out  defaults.Fonts
		errs []error
	)
	sum := fnv.New32a()
	for _, r := range Roles {
		path := Of(files, r)
		if path == "" {
			continue
		}
		f, err := Check(r, path)
		if err != nil {
			errs = append(errs, fmt.Errorf("font %s: %w", path, err))
			continue
		}
		out.Collection = append(out.Collection, f.Faces...)
		// The file as it is on disk: replaced by another, it measures anew.
		if info, err := os.Stat(path); err == nil {
			fmt.Fprintf(sum, "%d\x00%s\x00%d\x00%d\x00", r, path, info.Size(), info.ModTime().UnixNano())
		}
		switch r {
		case Text, Extra:
			out.Default = append(out.Default, f.Families...)
		case Mono:
			out.Preformatted = append(out.Preformatted, f.Families...)
		case Emoji:
			out.Emoji = f.Family
		}
	}
	rev := uint32(1)
	if len(out.Collection) > 0 {
		// 0 and 1 are taken.
		rev = sum.Sum32() | 2
	}
	revision.Store(rev)
	defaults.SetFonts(out)
	dropUnused(files)
	return errors.Join(errs...)
}

// dropUnused forgets the parsed files that no role uses any more, so that
// their memory is given back.
func dropUnused(files Files) {
	mu.Lock()
	defer mu.Unlock()
	for path := range cache {
		used := false
		for _, r := range Roles {
			used = used || Of(files, r) == path
		}
		if !used {
			delete(cache, path)
		}
	}
}
