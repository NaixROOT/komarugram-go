// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"testing"
	"time"

	"komarugram/internal/messenger/model"
)

// webmItem is a video sticker of the demo store, with its own media ID.
func webmItem(id string) model.PickerItem {
	return model.PickerItem{ID: id, Emoji: "🙂", Media: model.Message{Kind: model.MessageSticker,
		Media: &model.MessageMedia{ID: "demo/webm", MIMEType: "video/webm", Width: 512, Height: 512}}}
}

// played draws frames until the sticker of msg has decoded its loop.
func (h *composerHarness) played(t *testing.T, msg model.Message) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for !h.p.media.CheapToPlay(msg, image.Pt(1, 1), false) {
		if time.Now().After(deadline) {
			t.Fatal("the sticker did not decode its loop")
		}
		h.frame()
		time.Sleep(5 * time.Millisecond)
	}
}

// keepsStill reports whether msg shows its first frame without its loop.
func (h *composerHarness) keepsStill(msg model.Message) bool {
	return !h.p.media.CheapToPlay(msg, image.Pt(1, 1), false) && h.p.media.StatusFit(msg, false, image.Pt(1, 1), false).Frame != nil
}

// TestClosedStickerSetKeepsFirstFrames checks that the dialog closed drops
// its stickers' loops but keeps their first frames, which switching chats
// then forgets.
func TestClosedStickerSetKeepsFirstFrames(t *testing.T) {
	t.Chdir("../../..")
	h := newComposerHarness(t)
	h.animate = true
	item := webmItem("sticker/1")
	openStickerSet(h, model.StickerSet{Title: "Cats", Count: 1, Items: []model.PickerItem{item}})
	h.played(t, item.Media)
	h.p.stickers.modal.Close()
	for range 60 {
		h.frame()
	}
	if h.p.stickers.modal.Shown() {
		t.Fatal("the dialog did not close")
	}
	if !h.keepsStill(item.Media) {
		t.Fatal("the closed set did not keep only the first frame")
	}
	h.chat = 2
	h.frame()
	if h.p.media.StatusFit(item.Media, false, image.Pt(1, 1), false).Frame != nil {
		t.Fatal("switching chats kept the first frame of the set")
	}
}

// TestClosedPickerKeepsFirstFrames checks that the picker closed drops its
// stickers' loops but keeps their first frames, also in another chat.
func TestClosedPickerKeepsFirstFrames(t *testing.T) {
	t.Chdir("../../..")
	h := newComposerHarness(t)
	h.animate = true
	item := webmItem("sticker/1")
	c := h.p.composer
	c.tab = model.PickerStickers
	c.page = model.PickerPage{Packs: []model.PickerPack{{ID: 1, Items: []model.PickerItem{item}}}}
	c.selectedPack = 1
	c.pickerOpen = true
	h.played(t, item.Media)
	c.pickerOpen = false
	for range 60 {
		h.frame()
	}
	if !h.keepsStill(item.Media) {
		t.Fatal("the closed picker did not keep only the first frame")
	}
	// The same sticker in a set shown in the dialog of this chat.
	openStickerSet(h, model.StickerSet{Title: "Cats", Count: 1, Items: []model.PickerItem{item}})
	h.p.stickers.modal.Close()
	for range 60 {
		h.frame()
	}
	h.chat = 2
	h.frame()
	if !h.keepsStill(item.Media) {
		t.Fatal("switching chats forgot a first frame of the picker")
	}
}

// TestShownStickerViewsKeepMemory checks that the picker or the set dialog
// shown again cancels the release of the memory it dropped.
func TestShownStickerViewsKeepMemory(t *testing.T) {
	h := newComposerHarness(t)
	kept := 0
	h.p.keepMemory = func() { kept++ }
	c := h.p.composer
	for i := 1; i <= 2; i++ {
		c.pickerOpen = true
		for range 30 {
			h.frame()
		}
		if kept != i {
			t.Fatalf("the picker shown %d times kept the memory %d times", i, kept)
		}
		c.pickerOpen = false
		for range 60 {
			h.frame()
		}
	}
	h.p.stickers.open(h.p, model.StickerSetRef{Type: "id", ID: 1})
	if kept != 3 {
		t.Fatal("the set dialog shown did not keep the memory")
	}
}
