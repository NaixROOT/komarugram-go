// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"slices"

	"gio-mw/token"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
	"komarugram/internal/messenger/sendfiles"
)

// fileDrop is a drag of files from another program over the window, and
// the areas of the open chat they can be dropped on, as Telegram Desktop
// shows them: one for documents, and one for photos or media when the
// files are pictures or videos. Dropped, the files go to the box for
// sending files, the way the area says.
type fileDrop struct {
	// active is set while a drag is over the window; at is where, in the
	// window's content.
	active bool
	at     image.Point
	paths  []string
	state  sendfiles.DropState
	// page is the chat page the files go to, nil when none takes them,
	// and area where it is: both as the last frame drew them.
	page   *chatPage
	area   image.Rectangle
	metric unit.Metric
}

// Sizes of the areas, Telegram Desktop's dragMargin and dragPadding.
const (
	dropMarginY  = unit.Dp(10)
	dropPaddingX = unit.Dp(20)
	dropPaddingY = unit.Dp(10)
)

// Drop is called on the UI goroutine with each event of a drag of files
// over the window.
func (a *App) Drop(e app.DropEvent) { a.drop.take(e) }

// take follows a drag with one of its events, and gives the files dropped
// to the box for sending files of the page.
func (d *fileDrop) take(e app.DropEvent) {
	at := image.Pt(int(e.Position.X), int(e.Position.Y))
	if e.Paths != nil && !slices.Equal(e.Paths, d.paths) {
		d.paths, d.state = e.Paths, sendfiles.DropStateOf(e.Paths)
	}
	switch e.Kind {
	case app.DropEnter, app.DropMove:
		d.active, d.at = true, at
	case app.DropLeave:
		d.forget()
	case app.Drop:
		paths, state := d.paths, d.state
		page, documents := d.page, d.documentsAt(at, d.metric)
		d.forget()
		if page == nil || page.composer == nil || state == sendfiles.DropNone {
			return
		}
		// Dropped over the box, the files join it the way it sends.
		page.composer.files.addPaths(page.composer, paths, documents)
	}
}

// forget ends the drag, keeping where the chat is.
func (d *fileDrop) forget() {
	d.active, d.paths, d.state = false, nil, sendfiles.DropNone
}

// zones are the areas the files can be dropped on in d.area: the one for
// documents, and the one for photos or media, empty when there is none.
func (d *fileDrop) zones(m unit.Metric) (documents, media image.Rectangle) {
	area := d.area
	area.Min.Y += m.Dp(dropMarginY)
	area.Max.Y -= m.Dp(dropMarginY)
	if area.Dx() <= 0 || area.Dy() <= 0 {
		return image.Rectangle{}, image.Rectangle{}
	}
	switch d.state {
	case sendfiles.DropFiles:
		return area, image.Rectangle{}
	case sendfiles.DropPhotos, sendfiles.DropMedia:
		half := area.Dy() / 2
		documents = image.Rect(area.Min.X, area.Min.Y, area.Max.X, area.Min.Y+half)
		media = image.Rect(area.Min.X, area.Max.Y-half, area.Max.X, area.Max.Y)
		return documents, media
	}
	return image.Rectangle{}, image.Rectangle{}
}

// documentsAt reports whether files dropped at p go as documents: they do
// unless dropped on the area of photos or media.
func (d *fileDrop) documentsAt(p image.Point, m unit.Metric) bool {
	_, media := d.zones(m)
	return !p.In(media)
}

// takesFiles reports whether files may be sent to the chat of the page.
func (p *chatPage) takesFiles() bool {
	if p == nil || p.composer == nil || p.composer.source == nil || p.frozen.Frozen() {
		return false
	}
	return p.composer.permissions(p.chat).Any(model.SendAttachments)
}

// layout draws the areas of a drag over the chat, over the window, whose
// content gtx is.
func (d *fileDrop) layout(gtx layout.Context, l localization.Catalog) {
	if !d.active || d.page == nil || d.state == sendfiles.DropNone || d.page.composer.files.Shown() {
		return
	}
	documents, media := d.zones(gtx.Metric)
	var title, subtitle string
	switch d.state {
	case sendfiles.DropFiles, sendfiles.DropMedia:
		title, subtitle = l.T("drop.files_here"), l.T("drop.as_files")
	case sendfiles.DropPhotos:
		title, subtitle = l.T("drop.images_here"), l.T("drop.no_compression")
	}
	layoutDropZone(gtx, documents, title, subtitle, d.at.In(documents) && !d.at.In(media))
	switch d.state {
	case sendfiles.DropPhotos:
		layoutDropZone(gtx, media, l.T("drop.photos_here"), l.T("drop.quick"), d.at.In(media))
	case sendfiles.DropMedia:
		layoutDropZone(gtx, media, l.T("drop.media_here"), l.T("drop.as_media"), d.at.In(media))
	}
}

// layoutDropZone draws an area in zone: a card with what dropping there
// does, in the primary color while the drag is over it.
func layoutDropZone(gtx layout.Context, zone image.Rectangle, title, subtitle string, over bool) {
	inner := zone
	inner.Min.X += gtx.Dp(dropPaddingX)
	inner.Max.X -= gtx.Dp(dropPaddingX)
	inner.Min.Y += gtx.Dp(dropPaddingY)
	inner.Max.Y -= gtx.Dp(dropPaddingY)
	if inner.Dx() <= 0 || inner.Dy() <= 0 {
		return
	}
	sc := scheme(gtx)
	color, border := sc.SurfaceVariant.OnColor, sc.OutlineVariant
	if over {
		color, border = sc.Primary.Color, sc.Primary.Color
	}
	inRect(gtx, inner, func(gtx layout.Context) layout.Dimensions {
		size := gtx.Constraints.Max
		radius := gtx.Dp(16)
		fillRounded(gtx, sc.SurfaceContainerHigh, size, radius)
		width := float32(gtx.Dp(2))
		paint.FillShape(gtx.Ops, border.AsNRGBA(), clip.Stroke{Path: clip.UniformRRect(image.Rectangle{Max: size}, radius).Path(gtx.Ops), Width: width}.Op())
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = max(size.X-gtx.Dp(32), 0)
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return centeredLabel(gtx, title, token.TypestyleHeadlineSmall, color, 2)
				}),
				vspace(8),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return centeredLabel(gtx, subtitle, token.TypestyleTitleMedium, color, 2)
				}),
			)
		})
	})
}
