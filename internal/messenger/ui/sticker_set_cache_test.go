// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"errors"
	"image"
	"testing"
	"time"

	"komarugram/internal/messenger/mockstore"
	"komarugram/internal/messenger/model"
)

// cachedSets serves a set from its cache at once, and fetched when fetched
// is sent.
type cachedSets struct {
	*mockstore.Store
	cached  model.StickerSet
	fetched chan error
}

func (s *cachedSets) CachedStickerSet(context.Context, model.StickerSetRef) (model.StickerSet, bool) {
	return s.cached, true
}

func (s *cachedSets) StickerSet(ctx context.Context, ref model.StickerSetRef) (model.StickerSet, error) {
	select {
	case err := <-s.fetched:
		if err != nil {
			return model.StickerSet{}, err
		}
	case <-ctx.Done():
		return model.StickerSet{}, ctx.Err()
	}
	set := s.cached
	set.Title = "Fresh"
	return set, nil
}

func (s *cachedSets) SetStickerSetInstalled(context.Context, model.StickerSetRef, bool) error {
	return nil
}

// TestStickerSetShowsCachedSetFirst checks that the dialog shows a set as
// last fetched while fetching it, and keeps it when the fetch fails.
func TestStickerSetShowsCachedSetFirst(t *testing.T) {
	for _, fail := range []bool{false, true} {
		source := &cachedSets{Store: mockstore.New(time.Now(), 0), fetched: make(chan error, 1),
			cached: model.StickerSet{Title: "Cached", Count: 1, Items: []model.PickerItem{{ID: "sticker/1", Emoji: "🐈"}}}}
		p := newChatPage(source, func() {})
		p.images = &imageOps{}
		t.Cleanup(p.Close)
		h := &composerHarness{p: p, now: time.Now(), size: image.Pt(680, 720), chat: 1}
		h.frame()
		p.stickers.open(p, model.StickerSetRef{Type: "id", ID: 1})
		deadline := time.Now().Add(5 * time.Second)
		for p.stickers.busy || p.stickers.pack == nil {
			if time.Now().After(deadline) {
				t.Fatal("the cached set was not shown")
			}
			h.frame()
			time.Sleep(time.Millisecond)
		}
		if p.stickers.pack.Title != "Cached" {
			t.Fatalf("shows %q before the fetch", p.stickers.pack.Title)
		}
		var err error
		want := "Fresh"
		if fail {
			err, want = errors.New("offline"), "Cached"
		}
		source.fetched <- err
		for p.stickers.refreshing {
			if time.Now().After(deadline) {
				t.Fatal("the fetch did not end")
			}
			h.frame()
			time.Sleep(time.Millisecond)
		}
		if p.stickers.pack == nil || p.stickers.pack.Title != want || p.stickers.err != nil {
			t.Fatalf("failed %v: shows %+v, error %v", fail, p.stickers.pack, p.stickers.err)
		}
	}
}
