// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"komarugram/internal/messenger/model"
	"testing"
	"time"

	"github.com/gotd/td/tg"
)

func TestWebPreviewPhotoReferencesAndPlainMessage(t *testing.T) {
	s := testStore(t)
	raw := &tg.Message{ID: 5, PeerID: &tg.PeerUser{UserID: 2}, Message: "caption", Media: &tg.MessageMediaWebPage{Webpage: &tg.WebPage{URL: "https://example.org", Title: "From Telegram", Description: "Description", Photo: &tg.Photo{ID: 4, AccessHash: 6, FileReference: []byte{9}, Sizes: []tg.PhotoSizeClass{&tg.PhotoSize{Type: "m", W: 320, H: 240, Size: 42}}}}}}
	messages, e := s.ingest(context.Background(), []tg.MessageClass{raw}, false, 0)
	if e != nil {
		t.Fatal(e)
	}
	m := messages[0]
	if m.Media != nil || m.Kind != model.MessageText || m.WebPage == nil || m.WebPage.Photo == nil || m.WebPage.Title != "From Telegram" {
		t.Fatal(m)
	}
	if ref, ok := s.history.refs[m.WebPage.Photo.ID]; !ok || ref.Thumb != "m" || !ref.Photo {
		t.Fatal("Telegram preview is not downloadable", ref)
	}
}
func TestGiftAttributesPrivacyAndInstanceIdentity(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	peer := peerRecord{ID: 2, Kind: "user"}
	s.history.peers[7] = peerRecord{ID: 7, Kind: "user", Name: "Sender"}
	doc := &tg.Document{ID: 101, MimeType: "application/x-tgsticker"}
	pattern := &tg.Document{ID: 102, MimeType: "application/x-tgsticker"}
	raw := tg.SavedStarGift{Date: 100, MsgID: 9, FromID: &tg.PeerUser{UserID: 7}, Gift: &tg.StarGiftUnique{ID: 55, Title: "Gift", Num: 789, AvailabilityIssued: 100, AvailabilityTotal: 200, Attributes: []tg.StarGiftAttributeClass{
		&tg.StarGiftAttributeModel{Name: "Model", Document: doc, Rarity: &tg.StarGiftAttributeRarity{Permille: 25}}, &tg.StarGiftAttributePattern{Name: "Symbol", Document: pattern}, &tg.StarGiftAttributeBackdrop{Name: "Background", CenterColor: 0xabcdef, EdgeColor: 0x123456, PatternColor: 0x778899, TextColor: 0xffffff},
	}}}
	m, e := s.savedGiftMessage(ctx, peer, raw, "offset", 0)
	if e != nil {
		t.Fatal(e)
	}
	g := m.Gift
	if g == nil || g.Model != "Model" || g.Symbol != "Symbol" || g.Number != 789 || g.CenterColor != 0xabcdef || g.SenderName != "Sender" || g.ModelRarity != 25 || g.Pattern == nil {
		t.Fatal(g)
	}
	if ref := s.history.refs[g.Pattern.ID]; !ref.Gift || ref.GiftOffset != "offset" {
		t.Fatal("gift reference namespace missing", ref)
	}
	raw.NameHidden = true
	hidden, e := s.savedGiftMessage(ctx, peer, raw, "offset", 0)
	if e != nil || hidden.Gift.SenderID != 0 || hidden.Gift.SenderName != "" {
		t.Fatal("private sender leaked", hidden.Gift, e)
	}
	raw.MsgID = 10
	second, e := s.savedGiftMessage(ctx, peer, raw, "offset", 1)
	if e != nil || m.Key == second.Key {
		t.Fatal("gift instances collide")
	}
}
func TestLanguagePackKeepsPluralsAndLoadsOffline(t *testing.T) {
	values := languageStrings([]tg.LangPackStringClass{&tg.LangPackString{Key: "lng_media_type_photos", Value: "PHOTO"}, &tg.LangPackStringPluralized{Key: "lng_profile_photos", OneValue: "{count} one", FewValue: "{count} few", ManyValue: "{count} many", OtherValue: "{count} other"}})
	if values["lng_profile_photos#few"] != "{count} few" {
		t.Fatal(values)
	}
	s := testStore(t)
	if e := s.Cache().Put(context.Background(), "language/ru", values); e != nil {
		t.Fatal(e)
	}
	s.LanguagePack("ru")
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if got := s.LanguagePack("ru"); got["lng_media_type_photos"] == "PHOTO" && got["lng_profile_photos#many"] != "" {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("offline server language pack not restored")
}
