// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"fmt"
	"image"
	"sync"

	"komarugram/internal/crash"
	"komarugram/internal/messenger/chattheme"
	"komarugram/internal/messenger/model"

	"gio-mw/token"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
)

// wallpaperData is the picture of a wallpaper: its own bytes, the file of
// the settings load reads, or its media, the thumbnail when thumb is set
// and there is one.
func wallpaperData(ctx context.Context, media model.ConversationStore, w *model.ChatWallpaper, file string, load func(string) ([]byte, error), thumb bool) ([]byte, error) {
	switch {
	case len(w.Image) > 0:
		return w.Image, nil
	case file != "" && load != nil:
		return load(file)
	case w.Media != nil && media != nil:
		m := w.Media
		if thumb && m.Thumbnail != nil {
			m = m.Thumbnail
		}
		return media.Media(ctx, model.Message{Kind: model.MessagePhoto, Media: m})
	}
	return nil, nil
}

// wallpaperThumbs renders small wallpapers off the frame, for the cards of
// themes and wallpapers, and keeps them while they are drawn.
type wallpaperThumbs struct {
	invalidate func()
	media      model.ConversationStore
	ctx        context.Context
	cancel     context.CancelFunc
	// slots bounds the renders at once.
	slots   chan struct{}
	mu      sync.Mutex
	entries map[string]*thumbEntry
	frame   uint64
}

type thumbEntry struct {
	image *image.RGBA
	// average is the wallpaper's color, for plates over it.
	average uint32
	done    bool
	seen    uint64
}

func newWallpaperThumbs(media model.ConversationStore, invalidate func()) *wallpaperThumbs {
	ctx, cancel := context.WithCancel(context.Background())
	return &wallpaperThumbs{invalidate: invalidate, media: media, ctx: ctx, cancel: cancel, slots: make(chan struct{}, 3), entries: map[string]*thumbEntry{}}
}

// Frame forgets the thumbnails not drawn for a while; the owner calls it
// once a frame it draws them.
func (w *wallpaperThumbs) Frame() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.frame++
	for k, e := range w.entries {
		if e.done && e.seen+120 < w.frame {
			delete(w.entries, k)
		}
	}
}

// Release forgets every thumbnail.
func (w *wallpaperThumbs) Release() {
	w.mu.Lock()
	w.entries = map[string]*thumbEntry{}
	w.mu.Unlock()
}

// Close stops the renders.
func (w *wallpaperThumbs) Close() { w.cancel() }

// Get is wallpaper wp at size, in pixels, rendered from the picture load
// gives, or its own; nil until it is. key tells it from others, worked out
// from wp when empty: a caller with a large picture keeps its own.
func (w *wallpaperThumbs) Get(key string, wp *model.ChatWallpaper, size image.Point, load func(context.Context) ([]byte, error)) (*image.RGBA, uint32) {
	if key == "" {
		key = wallpaperKey(wp, "")
	}
	key = fmt.Sprintf("%s@%v", key, size)
	w.mu.Lock()
	defer w.mu.Unlock()
	e := w.entries[key]
	if e == nil {
		e = &thumbEntry{}
		w.entries[key] = e
		copy := *wp
		go w.render(e, &copy, size, load)
	}
	e.seen = w.frame
	return e.image, e.average
}

func (w *wallpaperThumbs) render(e *thumbEntry, wp *model.ChatWallpaper, size image.Point, load func(context.Context) ([]byte, error)) {
	defer crash.Recover("wallpaper thumbnail", nil)
	select {
	case w.slots <- struct{}{}:
	case <-w.ctx.Done():
		return
	}
	defer func() { <-w.slots }()
	var data []byte
	if load != nil {
		data, _ = load(w.ctx)
	} else {
		data, _ = wallpaperData(w.ctx, w.media, wp, "", nil, true)
	}
	// A picture that failed leaves the colors.
	b, _ := chattheme.Prepare(wp, data)
	im := b.Render(size)
	w.mu.Lock()
	e.image, e.average, e.done = im, b.Average(), true
	w.mu.Unlock()
	w.invalidate()
}

// drawCover fills size with im, cut to cover it.
func drawCover(gtx layout.Context, images *imageOps, im image.Image, size image.Point) {
	gtx.Constraints = layout.Exact(size)
	widget.Image{Src: images.Op(im), Fit: widget.Cover, Position: layout.Center}.Layout(gtx)
}

// themePreview draws a theme's look small, as Telegram Desktop's cards of
// themes: the wallpaper, im, with an incoming and an outgoing bubble over
// it; the application's colors for a nil style.
func themePreview(gtx layout.Context, images *imageOps, im image.Image, style *model.ChatThemeStyle, size image.Point, radius int) {
	sc := scheme(gtx)
	defer clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops).Pop()
	fillRect(gtx, sc.SurfaceContainerLow, size)
	if im != nil {
		drawCover(gtx, images, im, size)
	}
	in, out := sc.Surface.Color, sc.PrimaryContainer.Color
	if style != nil {
		in = token.NewMatColorFromHexRGBA(style.Incoming)
		if len(style.Outgoing) > 0 {
			out = token.NewMatColorFromHexRGB(style.Outgoing[0])
		}
	}
	w, h := size.X*11/20, max(size.Y/7, gtx.Dp(8))
	r := h / 2
	pad := size.X / 10
	top := size.Y / 5
	bubble := func(col token.MatColor, x, y int) {
		rect := image.Rect(x, y, x+w, y+h)
		paint.FillShape(gtx.Ops, col.AsNRGBA(), clip.UniformRRect(rect, r).Op(gtx.Ops))
	}
	bubble(in, pad, top)
	bubble(out, size.X-pad-w, top+h+h/2)
}
