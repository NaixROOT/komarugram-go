// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"gio-mw/token"

	"gioui.org/layout"
	"gioui.org/op"

	"komarugram/internal/messenger/fonts"
	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/preferences"
)

// fontFilter are the font files the shaper reads: TrueType and OpenType,
// and their collections.
var fontFilter = fileFilter{"Fonts", []string{".ttf", ".otf", ".ttc", ".otc"}}

// fontSettings is the part of the appearance settings that points the
// client at font files of the user's own: for the text, for the scripts
// that font lacks, for preformatted text and for emoji. A file is kept
// only once it loads as a font, and, for emoji, has them.
type fontSettings struct {
	files    func() preferences.Fonts
	setFiles func(preferences.Fonts)

	rows    [len(fonts.Roles)]fontRow
	results chan fontAnswer
	// invalidate redraws the window once a file is looked at.
	invalidate func()
}

// fontRow is one role's row: what is chosen, and the buttons.
type fontRow struct {
	choose, reset surface
	picking       bool
	// pickErr is why the file picked last was turned down.
	pickErr error
	// path is the file looked at last, and family its font's, or err why
	// it does not load.
	path, family string
	err          error
	known        bool
}

// fontAnswer is what a font file turned out to be.
type fontAnswer struct {
	role         fonts.Role
	path, family string
	err          error
	// picked marks a file the user picked now; cancelled, a chooser closed
	// without one.
	picked, cancelled bool
}

func newFontSettings(invalidate func()) *fontSettings {
	return &fontSettings{results: make(chan fontAnswer, 8), invalidate: invalidate}
}

func (s *fontSettings) Update(gtx layout.Context) {
	if s.files == nil || s.setFiles == nil {
		return
	}
	for done := false; !done; {
		select {
		case a := <-s.results:
			row := &s.rows[a.role]
			if a.picked {
				row.picking = false
				if a.cancelled {
					break
				}
				row.pickErr = a.err
				if a.err == nil {
					s.setFiles(fonts.With(s.files(), a.role, a.path))
					row.path, row.family, row.err, row.known = a.path, a.family, nil, true
				}
			} else if a.path == row.path {
				row.family, row.err, row.known = a.family, a.err, true
			}
		default:
			done = true
		}
	}
	files := s.files()
	for _, role := range fonts.Roles {
		row := &s.rows[role]
		// The saved file is looked at once, off the frame: a large font
		// takes a moment to read.
		if path := fonts.Of(files, role); path != row.path {
			row.path, row.family, row.err, row.known = path, "", nil, path == ""
			if path != "" {
				go s.look(fontAnswer{role: role, path: path})
			}
		}
		if row.choose.Clicked(gtx) && !row.picking {
			row.picking, row.pickErr = true, nil
			go s.pick(role)
		}
		if row.reset.Clicked(gtx) {
			row.pickErr = nil
			s.setFiles(fonts.With(files, role, ""))
			files = s.files()
			gtx.Execute(op.InvalidateCmd{})
		}
	}
}

// look loads the file of a and answers what it is.
func (s *fontSettings) look(a fontAnswer) {
	f, err := fonts.Check(a.role, a.path)
	a.family, a.err = f.Family, err
	s.results <- a
	s.invalidate()
}

// pick lets the user choose a font file and looks at it.
func (s *fontSettings) pick(role fonts.Role) {
	choice := chooseFile(context.Background(), &fontFilter)
	a := fontAnswer{role: role, path: choice.path, err: choice.err, picked: true}
	if choice.err == nil && choice.path == "" {
		a.cancelled = true
		s.results <- a
		s.invalidate()
		return
	}
	if a.err != nil {
		s.results <- a
		s.invalidate()
		return
	}
	s.look(a)
}

// applyFonts makes the themes use the font files of the settings. A file
// that is gone or does not load is passed over: the settings say so.
func applyFonts(f preferences.Fonts) {
	if err := fonts.Apply(f); err != nil {
		log.Printf("fonts: %v", err)
	}
}

// fontRoleKeys are the texts of the roles: the title and what it is for.
var fontRoleKeys = map[fonts.Role][2]string{
	fonts.Text:  {"fonts.text", "fonts.text_hint"},
	fonts.Extra: {"fonts.extra", "fonts.extra_hint"},
	fonts.Mono:  {"fonts.mono", "fonts.mono_hint"},
	fonts.Emoji: {"fonts.emoji", "fonts.emoji_hint"},
}

func (s *fontSettings) Layout(gtx layout.Context, l localization.Catalog) layout.Dimensions {
	sc := scheme(gtx)
	rows := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("fonts.title"), token.TypestyleTitleMedium, sc.Surface.OnColor, 1)
		}),
		vspace(4),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("fonts.body"), token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 0)
		}),
	}
	files := s.files()
	for _, role := range fonts.Roles {
		rows = append(rows, vspace(12), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutRow(gtx, role, fonts.Of(files, role), l)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

func (s *fontSettings) layoutRow(gtx layout.Context, role fonts.Role, path string, l localization.Catalog) layout.Dimensions {
	sc := scheme(gtx)
	row := &s.rows[role]
	keys := fontRoleKeys[role]
	line := func(text string, failed bool) layout.FlexChild {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			color := sc.SurfaceVariant.OnColor
			if failed {
				color = sc.Error.Color
			}
			return label(gtx, text, token.TypestyleBodyMedium, color, 3)
		})
	}
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T(keys[0]), token.TypestyleTitleSmall, sc.Surface.OnColor, 1)
		}),
		line(l.T(keys[1]), false),
	}
	env := fonts.FromEnv(role)
	switch {
	case env:
		children = append(children, line(l.T("fonts.from_env"), false))
	case path == "":
		children = append(children, line(l.T("fonts.system"), false))
	case !row.known || row.path != path:
		children = append(children, line(filepath.Base(path), false))
	case row.err != nil:
		children = append(children, line(filepath.Base(path)+": "+fontErrorText(row.err, l), true))
	default:
		children = append(children, line(row.family+" · "+filepath.Base(path), false))
	}
	if row.pickErr != nil {
		children = append(children, line(l.T("fonts.rejected")+": "+fontErrorText(row.pickErr, l), true))
	}
	if !env {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			buttons := []layout.FlexChild{layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if row.picking {
					gtx = gtx.Disabled()
				}
				return textButton(gtx, &row.choose, l.T("fonts.choose"))
			})}
			if path != "" {
				buttons = append(buttons, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return textButton(gtx, &row.reset, l.T("fonts.reset"))
				}))
			}
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx, buttons...)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

// fontErrorText says why a font file was turned down.
func fontErrorText(err error, l localization.Catalog) string {
	switch {
	case errors.Is(err, fonts.ErrNoEmoji):
		return l.T("fonts.no_emoji")
	case errors.Is(err, fonts.ErrColorFormat):
		return l.T("fonts.color_format")
	case errors.Is(err, fonts.ErrTooLarge):
		return l.T("fonts.too_large")
	case errors.Is(err, os.ErrNotExist):
		return l.T("fonts.missing")
	}
	return l.T("fonts.not_a_font")
}
