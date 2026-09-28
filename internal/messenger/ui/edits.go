// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"image"
	"time"

	"gio-mw/token"
	"gio-mw/widget/scroll"

	"gioui.org/layout"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

// editsDialog shows the versions an edited message had, as AyuGram's
// edits history does: oldest first, then the message as it is.
type editsDialog struct {
	modal    modal
	msg      model.Message
	versions []model.Message
	loaded   bool
	failed   bool
	results  chan editsResult
	cancel   context.CancelFunc
	close    surface
	list     scroll.List
	loader   loadingIndicator
}

type editsResult struct {
	versions []model.Message
	err      error
}

// open shows the versions of m, which load in the background.
func (d *editsDialog) open(p *chatPage, m model.Message) {
	d.stop()
	store, ok := p.source.(model.KeepStore)
	if !ok {
		return
	}
	d.msg = m
	d.modal.Open()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	d.cancel = cancel
	d.results = make(chan editsResult, 1)
	results := d.results
	go func() {
		defer cancel()
		versions, err := store.MessageEdits(ctx, m)
		results <- editsResult{versions, err}
		p.invalidate()
	}()
}

func (d *editsDialog) stop() {
	if d.cancel != nil {
		d.cancel()
	}
	*d = editsDialog{}
}

func (d *editsDialog) layout(gtx layout.Context, p *chatPage, l localization.Catalog) {
	if !d.modal.Shown() {
		return
	}
	select {
	case r := <-d.results:
		d.loaded, d.failed, d.versions = true, r.err != nil, r.versions
	default:
	}
	if !d.modal.closing && d.close.Clicked(gtx) {
		d.modal.Close()
	}
	shown := d.modal.Layout(gtx, false, func(gtx layout.Context) layout.Dimensions {
		width := min(gtx.Constraints.Max.X, gtx.Dp(460))
		height := min(gtx.Constraints.Max.Y, gtx.Dp(520))
		gtx.Constraints = layout.Exact(image.Pt(width, height))
		return card(gtx, func(gtx layout.Context) layout.Dimensions {
			sc := scheme(gtx)
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return label(gtx, l.T("edits.title"), token.TypestyleTitleLarge, sc.Surface.OnColor, 1)
				}),
				vspace(12),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					switch {
					case !d.loaded:
						return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return d.loader.sized(gtx, l, 32) })
					case d.failed:
						return label(gtx, l.T("edits.failed"), token.TypestyleBodyMedium, sc.Error.Color, 2)
					case len(d.versions) == 0:
						return label(gtx, l.T("edits.none"), token.TypestyleBodyMedium, sc.SurfaceVariant.OnColor, 0)
					}
					all := append(append([]model.Message(nil), d.versions...), d.msg)
					d.list.Axis = layout.Vertical
					return d.list.Layout(gtx, len(all), func(gtx layout.Context, i int) layout.Dimensions {
						return d.version(gtx, all[i], i == len(all)-1, l)
					})
				}),
				vspace(12),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return textButton(gtx, &d.close, l.T("stickers.close"))
					})
				}),
			)
		}, defaultCardPadding)
	})
	if !shown {
		d.stop()
	}
}

// version draws one version: when it was written or edited, and its text.
func (d *editsDialog) version(gtx layout.Context, m model.Message, current bool, l localization.Catalog) layout.Dimensions {
	sc := scheme(gtx)
	at := m.EditedAt
	if at.IsZero() {
		at = m.Date
	}
	when := at.Local().Format("02.01.2006 15:04")
	if current {
		when += " · " + l.T("edits.current")
	}
	return layout.Inset{Bottom: 12}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return label(gtx, when, token.TypestyleLabelMedium, sc.Primary.Color, 1)
			}),
			vspace(2),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return label(gtx, m.Text, token.TypestyleBodyMedium, sc.Surface.OnColor, 0)
			}),
		)
	})
}
