// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"gio-mw/token"

	"gioui.org/layout"
	"gioui.org/op"

	"komarugram/internal/messenger/emojipacks"
	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/preferences"
)

// emojiSettings is the part of the appearance settings that chooses what
// draws emoji: a pack of the catalog, downloaded once and kept beside the
// settings, or none, which leaves them to the fonts. A pack can be
// downloaded, chosen, downloaded again when the catalog has a newer one,
// and deleted.
type emojiSettings struct {
	files    func() preferences.Fonts
	setFiles func(preferences.Fonts)
	// store keeps the packs; source is the catalog, nil for none.
	store  *emojipacks.Store
	source emojipacks.Source
	// applied is told that the files of the pack in use changed.
	applied func()

	// catalog are the packs of the catalog, once it is read; catalogErr is
	// why it was not.
	catalog    []emojipacks.Pack
	catalogErr error
	asked      bool
	reading    bool

	// have are the packs installed, as read from the disk at haveAt: every
	// frame asks, and another window may install one.
	have   []emojipacks.Pack
	haveAt time.Time

	none   surface
	rows   map[string]*emojiPackRow
	events chan emojiPackEvent
	// invalidate redraws the window once something came.
	invalidate func()
}

// emojiPackRow is the row of a pack.
type emojiPackRow struct {
	use, download, remove, cancel surface
	// stop ends the download going on; nil without one.
	stop        context.CancelFunc
	done, total int64
	// err is why the pack was not downloaded or deleted last.
	err error
}

// emojiPackEvent is what a download or the catalog came to.
type emojiPackEvent struct {
	// id is the pack of a download's progress or end; "" for the catalog.
	id          string
	done, total int64
	finished    bool
	err         error
	catalog     []emojipacks.Pack
}

// emojiPackDir is where the emoji packs are kept: beside the settings, or,
// for settings kept in memory, as the demo's are, in the system's temporary
// directory, so that packs can be tried there too.
func emojiPackDir(prefs *preferences.Store) string {
	if dir := prefs.DataDir("emoji"); dir != "" {
		return dir
	}
	return filepath.Join(os.TempDir(), "komarugram-go-demo", "emoji")
}

func newEmojiSettings(invalidate func()) *emojiSettings {
	return &emojiSettings{rows: map[string]*emojiPackRow{}, events: make(chan emojiPackEvent, 64), invalidate: invalidate}
}

// available reports whether there is anything to choose from: a catalog,
// or packs installed before.
func (s *emojiSettings) available() bool {
	return s.files != nil && s.setFiles != nil && s.store != nil && (s.source != nil || len(s.installed()) > 0)
}

// installed are the packs installed, read again once a second and after
// what these settings did to them.
func (s *emojiSettings) installed() []emojipacks.Pack {
	if now := time.Now(); s.haveAt.IsZero() || now.Sub(s.haveAt) > time.Second {
		s.have, s.haveAt = s.store.Installed(), now
	}
	return s.have
}

func (s *emojiSettings) row(id string) *emojiPackRow {
	row := s.rows[id]
	if row == nil {
		row = new(emojiPackRow)
		s.rows[id] = row
	}
	return row
}

// packs are the packs shown: those of the catalog, and the installed ones
// the catalog does not have, by name.
func (s *emojiSettings) packs() (shown []emojipacks.Pack, installed map[string]emojipacks.Pack) {
	installed = map[string]emojipacks.Pack{}
	for _, p := range s.installed() {
		installed[p.ID] = p
	}
	inCatalog := map[string]bool{}
	for _, p := range s.catalog {
		inCatalog[p.ID] = true
		shown = append(shown, p)
	}
	for id, p := range installed {
		if !inCatalog[id] {
			shown = append(shown, p)
		}
	}
	sort.SliceStable(shown, func(i, j int) bool { return shown[i].Name < shown[j].Name })
	return shown, installed
}

func (s *emojiSettings) Update(gtx layout.Context) {
	if !s.available() {
		return
	}
	for done := false; !done; {
		select {
		case e := <-s.events:
			switch {
			case e.id == "":
				s.reading, s.catalog, s.catalogErr = false, e.catalog, e.err
			case e.finished:
				row := s.row(e.id)
				row.stop, row.err, s.haveAt = nil, e.err, time.Time{}
				if errors.Is(e.err, context.Canceled) {
					row.err = nil
				}
				// The pack in use is drawn from its new files.
				if e.err == nil && s.files().EmojiPack == e.id && s.applied != nil {
					s.applied()
				}
			default:
				row := s.row(e.id)
				row.done, row.total = e.done, e.total
			}
		default:
			done = true
		}
	}
	if s.source != nil && !s.asked {
		s.asked, s.reading = true, true
		go s.readCatalog()
	}
	files := s.files()
	if s.none.Clicked(gtx) && files.EmojiPack != "" {
		files.EmojiPack = ""
		s.setFiles(files)
		gtx.Execute(op.InvalidateCmd{})
	}
	shown, installed := s.packs()
	for _, p := range shown {
		row := s.row(p.ID)
		_, has := installed[p.ID]
		if row.use.Clicked(gtx) && has && files.EmojiPack != p.ID {
			files.EmojiPack = p.ID
			s.setFiles(files)
			gtx.Execute(op.InvalidateCmd{})
		}
		if row.download.Clicked(gtx) && row.stop == nil && s.source != nil {
			ctx, stop := context.WithCancel(context.Background())
			row.stop, row.err, row.done, row.total = stop, nil, 0, p.DownloadSize()
			go s.install(ctx, p)
		}
		if row.cancel.Clicked(gtx) && row.stop != nil {
			row.stop()
		}
		if row.remove.Clicked(gtx) && has && row.stop == nil {
			// The fonts let the pack go before its files do.
			if files.EmojiPack == p.ID {
				files.EmojiPack = ""
				s.setFiles(files)
			}
			row.err, s.haveAt = s.store.Remove(p.ID), time.Time{}
			gtx.Execute(op.InvalidateCmd{})
		}
	}
}

func (s *emojiSettings) readCatalog() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	packs, err := emojipacks.ReadIndex(ctx, s.source)
	s.events <- emojiPackEvent{catalog: packs, err: err}
	s.invalidate()
}

// install downloads a pack, telling of its progress a few times a second.
func (s *emojiSettings) install(ctx context.Context, p emojipacks.Pack) {
	var told time.Time
	err := s.store.Install(ctx, s.source, p, func(done, total int64) {
		if now := time.Now(); now.Sub(told) > 100*time.Millisecond {
			told = now
			select {
			case s.events <- emojiPackEvent{id: p.ID, done: done, total: total}:
				s.invalidate()
			default:
			}
		}
	})
	s.events <- emojiPackEvent{id: p.ID, finished: true, err: err}
	s.invalidate()
}

func megabytes(l localization.Catalog, size int64) string {
	return l.Format("fonts.size", map[string]string{"size": strconv.FormatFloat(float64(size)/(1<<20), 'f', 1, 64)})
}

func (s *emojiSettings) Layout(gtx layout.Context, l localization.Catalog) layout.Dimensions {
	sc := scheme(gtx)
	files := s.files()
	shown, installed := s.packs()
	rows := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("emojipacks.title"), token.TypestyleTitleMedium, sc.Surface.OnColor, 1)
		}),
		vspace(4),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("emojipacks.body"), token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 0)
		}),
		vspace(12),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutRow(gtx, l.T("emojipacks.none"), []string{l.T("emojipacks.none_hint")}, nil, files.EmojiPack == "", &s.none, nil, l)
		}),
	}
	for _, p := range shown {
		rows = append(rows, vspace(12), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			row := s.row(p.ID)
			have, has := installed[p.ID]
			kind := l.T("emojipacks.font")
			if p.Kind == emojipacks.KindSprites {
				kind = l.T("emojipacks.sprites")
			}
			about := kind + " · " + megabytes(l, p.DownloadSize())
			if p.License != "" {
				about += " · " + p.License
			}
			lines := []string{about}
			inCatalog := false
			for _, c := range s.catalog {
				inCatalog = inCatalog || c.ID == p.ID
			}
			outdated := has && inCatalog && have.Revision() != p.Revision()
			switch {
			case row.stop != nil:
				percent := 0
				if row.total > 0 {
					percent = int(row.done * 100 / row.total)
				}
				lines = append(lines, l.Format("emojipacks.progress", map[string]string{"percent": strconv.Itoa(percent)}))
			case outdated:
				lines = append(lines, l.T("emojipacks.update_available"))
			case has && !inCatalog && s.source != nil && !s.reading && s.catalogErr == nil:
				lines = append(lines, l.T("emojipacks.not_in_catalog"))
			}
			var failed []string
			switch {
			case errors.Is(row.err, emojipacks.ErrNeedsTelegram):
				failed = append(failed, l.T("emojipacks.needs_telegram"))
			case row.err != nil:
				failed = append(failed, l.T("emojipacks.failed")+": "+mediaErrorText(row.err))
			}
			var buttons []emojiPackButton
			switch {
			case row.stop != nil:
				buttons = append(buttons, emojiPackButton{&row.cancel, l.T("emojipacks.cancel")})
			case !has:
				buttons = append(buttons, emojiPackButton{&row.download, l.T("emojipacks.download")})
			default:
				if outdated {
					buttons = append(buttons, emojiPackButton{&row.download, l.T("emojipacks.update")})
				}
				buttons = append(buttons, emojiPackButton{&row.remove, l.T("emojipacks.remove")})
			}
			var use *surface
			if has {
				use = &row.use
			}
			return s.layoutRow(gtx, p.Name, lines, failed, has && files.EmojiPack == p.ID, use, buttons, l)
		}))
	}
	switch {
	case s.reading:
		rows = append(rows, vspace(12), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("emojipacks.reading"), token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 1)
		}))
	case s.catalogErr != nil:
		rows = append(rows, vspace(12), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, l.T("emojipacks.catalog_failed")+": "+mediaErrorText(s.catalogErr), token.TypestyleBodyMedium, sc.Error.Color, 3)
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, rows...)
}

// emojiPackButton is a button of a pack's row.
type emojiPackButton struct {
	click *surface
	text  string
}

// layoutRow draws a choice: its title, the lines about it, those of what
// failed, and its buttons. use chooses it, unless it is chosen already or
// cannot be.
func (s *emojiSettings) layoutRow(gtx layout.Context, title string, lines, failed []string, chosen bool, use *surface, buttons []emojiPackButton, l localization.Catalog) layout.Dimensions {
	sc := scheme(gtx)
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, title, token.TypestyleTitleSmall, sc.Surface.OnColor, 1)
		}),
	}
	for _, text := range lines {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, text, token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 3)
		}))
	}
	for _, text := range failed {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return label(gtx, text, token.TypestyleBodyMedium, sc.Error.Color, 3)
		}))
	}
	children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		var row []layout.FlexChild
		switch {
		case chosen:
			row = append(row, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{Top: 7, Bottom: 7, Left: 8, Right: 8}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return label(gtx, l.T("emojipacks.used"), token.TypestyleLabelLarge, sc.Surface.OnColor, 1)
				})
			}))
		case use != nil:
			row = append(row, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return textButton(gtx, use, l.T("emojipacks.use"))
			}))
		}
		for _, b := range buttons {
			row = append(row, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return textButton(gtx, b.click, b.text)
			}))
		}
		return layout.Flex{Alignment: layout.Middle}.Layout(gtx, row...)
	}))
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}
