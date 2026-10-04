// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"slices"
	"testing"

	"gioui.org/io/key"
	"gioui.org/layout"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

// A video in the gallery shows its thumbnail, never downloads the video,
// and plays in the external player by a click.
func TestPhotoViewerPlaysVideos(t *testing.T) {
	h := newViewerHarness(t)
	video := &h.store.photos[1]
	video.Kind = model.MessageVideo
	video.Media = &model.MessageMedia{ID: "v20/file", Width: 1280, Height: 720, Thumbnail: &model.MessageMedia{ID: "v20/t", Width: 320, Height: 180}}
	var played []model.MessageID
	h.viewer.play = func(_ layout.Context, m model.Message, _ localization.Catalog) {
		played = append(played, m.Key.MessageID)
	}

	h.viewer.Open(1, h.store.photos[0], nil)
	h.until("the video in the gallery", func() bool {
		items, _, _, _, _ := h.viewer.snapshot()
		return indexOf(items, 20) >= 0
	})
	h.key(key.NameRightArrow)
	h.until("the thumbnail", func() bool { return slices.Contains(h.store.read(), "v20/t") })
	if canKeep(*video) {
		t.Fatal("a video offers save and copy")
	}
	h.click(400, 300)
	if !slices.Equal(played, []model.MessageID{20}) {
		t.Fatalf("played %v", played)
	}
	if slices.Contains(h.store.read(), "v20/file") {
		t.Fatal("the video was downloaded")
	}
}

// A GIF opens alone: no gallery is read around it.
func TestPhotoViewerOpensGIFAlone(t *testing.T) {
	h := newViewerHarness(t)
	gif := model.Message{Key: model.MessageKey{ChatID: 1, MessageID: 25}, Kind: model.MessageGIF, Media: &model.MessageMedia{ID: "g25", MIMEType: "image/gif", Width: 200, Height: 100}}
	h.viewer.OpenAlone(1, gif)
	for range 5 {
		h.frame()
	}
	items, _, _, _, _ := h.viewer.snapshot()
	if len(items) != 1 || items[0].Key.MessageID != 25 || !h.viewer.open {
		t.Fatalf("%d items, open %v", len(items), h.viewer.open)
	}
}

// A link of a caption opens, through the chat's question, once the viewer
// closes; the rest of the caption is plain text.
func TestPhotoViewerCaptionLinks(t *testing.T) {
	h := newViewerHarness(t)
	photo := &h.store.photos[0]
	photo.Text = "see telegram.org now"
	photo.Entities = []model.Entity{{Kind: "url", Offset: 4, Length: 12}}
	var opened []string
	h.viewer.openLink = func(url string) { opened = append(opened, url) }
	h.viewer.Open(1, *photo, nil)
	h.frame()
	c := &h.viewer.caption
	var link image.Rectangle
	for _, f := range c.fragments {
		if c.runs[f.Index].URL != "" {
			link = f.Bounds.Add(c.origin)
		}
	}
	if link.Empty() {
		t.Fatalf("no link drawn: %+v", c.runs)
	}
	plain := c.fragments[0].Bounds.Add(c.origin)
	h.click(float32(plain.Min.X+2), float32(plain.Min.Y+plain.Dy()/2))
	if len(opened) != 0 || !h.viewer.open {
		t.Fatalf("plain text opened %v", opened)
	}
	h.click(float32(link.Min.X+link.Dx()/2), float32(link.Min.Y+link.Dy()/2))
	if !slices.Equal(opened, []string{"telegram.org"}) || h.viewer.open {
		t.Fatalf("opened %v, viewer open %v", opened, h.viewer.open)
	}
}

// A profile photo's video plays over it while animations are on.
func TestPhotoViewerPlaysProfileVideo(t *testing.T) {
	h := newViewerHarness(t)
	photo := h.store.photos[0]
	photo.Media = &model.MessageMedia{ID: "p10/w", Width: 1600, Height: 1200, Video: &model.MessageMedia{ID: "p10/video", MIMEType: "video/mp4", Width: 800, Height: 800}}
	h.viewer.Open(1, photo, nil)
	for range 5 {
		h.frame()
	}
	if slices.Contains(h.store.read(), "p10/video") {
		t.Fatal("the video was read with animations off")
	}
	h.animate = true
	h.until("the video", func() bool { return slices.Contains(h.store.read(), "p10/video") })
}
