// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"testing"

	"komarugram/internal/messenger/model"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func TestStickerSetReferenceSurvivesHistoryCache(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	raw := &tg.Message{
		ID: 18, PeerID: &tg.PeerUser{UserID: 5},
		Media: &tg.MessageMediaDocument{Document: &tg.Document{
			ID: 42, MimeType: "image/webp",
			Attributes: []tg.DocumentAttributeClass{&tg.DocumentAttributeSticker{
				Stickerset: &tg.InputStickerSetID{ID: 123, AccessHash: 456},
			}},
		}},
	}
	messages, err := s.ingest(ctx, []tg.MessageClass{raw}, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	check := func(m model.Message) {
		t.Helper()
		if m.Media == nil || m.Media.StickerSet == nil || m.Media.StickerSet.Type != "id" || m.Media.StickerSet.ID != 123 || m.Media.StickerSet.AccessHash != 456 {
			t.Fatalf("sticker set reference lost: %+v", m.Media)
		}
	}
	check(messages[0])
	cached, err := s.Cache().Around(ctx, messages[0].Key.ChatID, 0, 10)
	if err != nil || len(cached) != 1 {
		t.Fatalf("cached messages: %d, %v", len(cached), err)
	}
	check(cached[0])
}

func TestStickerSetFetchAndInstallRequests(t *testing.T) {
	s := testStore(t)
	set := tg.StickerSet{ID: 123, AccessHash: 456, Title: "Cats", Count: 1}
	set.Emojis = true
	set.SetInstalledDate(1)
	doc := &tg.Document{ID: 42, MimeType: "image/webp", Attributes: []tg.DocumentAttributeClass{
		&tg.DocumentAttributeSticker{Alt: "🐈", Stickerset: &tg.InputStickerSetID{ID: 123, AccessHash: 456}},
	}}
	var requests []string
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		switch req := in.(type) {
		case *tg.MessagesGetStickerSetRequest:
			ref, ok := req.Stickerset.(*tg.InputStickerSetID)
			if !ok || ref.ID != 123 || ref.AccessHash != 456 || req.Hash != 0 {
				t.Fatalf("get reference: %+v", req)
			}
			requests = append(requests, "get")
			return &tg.MessagesStickerSet{Set: set, Documents: []tg.DocumentClass{doc}}, nil
		case *tg.MessagesInstallStickerSetRequest:
			if req.Archived {
				t.Fatal("install archived set")
			}
			if ref, ok := req.Stickerset.(*tg.InputStickerSetID); !ok || ref.ID != 123 || ref.AccessHash != 456 {
				t.Fatalf("install reference: %+v", req.Stickerset)
			}
			requests = append(requests, "install")
			return &tg.MessagesStickerSetInstallResultSuccess{}, nil
		case *tg.MessagesUninstallStickerSetRequest:
			if ref, ok := req.Stickerset.(*tg.InputStickerSetID); !ok || ref.ID != 123 || ref.AccessHash != 456 {
				t.Fatalf("uninstall reference: %+v", req.Stickerset)
			}
			requests = append(requests, "uninstall")
			return &tg.BoolTrue{}, nil
		default:
			return nil, fmt.Errorf("unexpected %T", in)
		}
	})
	ref := model.StickerSetRef{Type: "id", ID: 123, AccessHash: 456}
	pack, err := s.StickerSet(context.Background(), ref)
	if err != nil || pack.Title != "Cats" || pack.Count != 1 || len(pack.Items) != 1 || pack.Items[0].Media.Media.StickerSet.ID != 123 || !pack.Installed || !pack.Emoji {
		t.Fatalf("fetched set: %+v, %v", pack, err)
	}
	if err = s.SetStickerSetInstalled(context.Background(), pack.Ref, true); err != nil {
		t.Fatal(err)
	}
	if err = s.SetStickerSetInstalled(context.Background(), pack.Ref, false); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(requests) != "[get install uninstall]" {
		t.Fatal(requests)
	}
}

func TestCustomEmojiDocumentKeepsStickerSet(t *testing.T) {
	for _, tc := range []struct {
		name      string
		set       tg.InputStickerSetClass
		typeName  string
		shortName string
	}{
		{"short name", &tg.InputStickerSetShortName{ShortName: "custom_pack"}, "short_name", "custom_pack"},
		{"default statuses", &tg.InputStickerSetEmojiDefaultStatuses{}, "emoji_default_statuses", ""},
		{"empty", &tg.InputStickerSetEmpty{}, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, media, _ := documentMedia(&tg.Document{ID: 7, Attributes: []tg.DocumentAttributeClass{
				&tg.DocumentAttributeCustomEmoji{Stickerset: tc.set},
			}})
			if tc.typeName == "" {
				if media.StickerSet != nil {
					t.Fatalf("empty set: %+v", media.StickerSet)
				}
				return
			}
			if media.StickerSet == nil || media.StickerSet.Type != tc.typeName || media.StickerSet.ShortName != tc.shortName {
				t.Fatalf("custom emoji set: %+v", media.StickerSet)
			}
		})
	}
}
