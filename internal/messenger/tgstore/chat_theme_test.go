// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"komarugram/internal/messenger/model"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func TestCloudAppearancePortableColorsAndWallpaperOverride(t *testing.T) {
	s := testStore(t)
	s.history.peers[42] = peerRecord{ID: 42, Hash: 7, Kind: "user"}
	bg := tg.WallPaperSettings{}
	bg.SetBackgroundColor(0)
	bg.SetSecondBackgroundColor(0x123456)
	bg.SetIntensity(-40)
	themeCalls := 0
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
		switch in.(type) {
		case *tg.UsersGetFullUserRequest:
			out.(*tg.UsersUserFull).FullUser = tg.UserFull{Theme: &tg.ChatTheme{Emoticon: "🌿"}, Wallpaper: &tg.WallPaperNoFile{Settings: bg}}
		case *tg.AccountGetChatThemesRequest:
			themeCalls++
			out.(*tg.AccountThemesBox).Themes = &tg.AccountThemes{Themes: []tg.Theme{{Emoticon: "🌿", Title: "Garden", Settings: []tg.ThemeSettings{{BaseTheme: &tg.BaseThemeDay{}, AccentColor: 0xff112233, MessageColors: []int{0x102030, 0x304050}}, {BaseTheme: &tg.BaseThemeNight{}, AccentColor: 0xff8899aa, MessageColors: []int{0x203040}}}}}}
		default:
			t.Fatalf("unexpected %T", in)
		}
		return nil
	}))
	a, err := s.ChatAppearance(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if a.Theme.Light == nil || a.Theme.Dark == nil || !a.Theme.Dark.Dark || a.Theme.Light.Accent != 0x112233 || a.Theme.Light.OutText != 0xffffffff {
		t.Fatal(a.Theme)
	}
	if len(a.Wallpaper.Colors) != 2 || a.Wallpaper.Colors[0] != 0 || a.Wallpaper.Intensity != -40 {
		t.Fatal("black color flag or wallpaper override lost", a.Wallpaper)
	}
	_, err = s.ChatAppearance(context.Background(), 42)
	if err != nil || themeCalls != 1 {
		t.Fatal("catalogue not cached", themeCalls, err)
	}
	s.history.api = nil
	cached, err := s.ChatAppearance(context.Background(), 42)
	if err != nil || cached.Theme.ID != "🌿" {
		t.Fatal("offline appearance", cached, err)
	}
}
func TestThemeLocalOverrideAndServerApply(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	s.history.peers[42] = peerRecord{ID: 42, Kind: "user"}
	local := &model.ChatTheme{Title: "Local", Light: &model.ChatThemeStyle{Incoming: 0x112233ff, Wallpaper: &model.ChatWallpaper{Image: []byte("wallpaper")}}}
	if err := s.SetLocalChatTheme(ctx, 42, local); err != nil {
		t.Fatal(err)
	}
	got, err := s.ChatAppearance(ctx, 42)
	if err != nil || got.Theme.Title != "Local" {
		t.Fatal(got, err)
	}
	called := false
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
		req := in.(*tg.MessagesSetChatThemeRequest)
		if req.Theme.(*tg.InputChatTheme).Emoticon != "🌿" {
			t.Fatal(req.Theme)
		}
		called = true
		out.(*tg.UpdatesBox).Updates = &tg.Updates{}
		return nil
	}))
	if err := s.SetChatTheme(ctx, 42, "🌿"); err != nil || !called {
		t.Fatal(err, called)
	}
	var saved *model.ChatTheme
	if _, err := s.Cache().Get(ctx, "local-theme/42", &saved); err != nil || saved != nil {
		t.Fatal("server theme hidden by local override", saved, err)
	}
}
func TestStoriesStayOutsideHistoryAndKeepReferences(t *testing.T) {
	s := testStore(t)
	s.history.peers[42] = peerRecord{ID: 42, Kind: "user"}
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
		req := in.(*tg.StoriesGetPinnedStoriesRequest)
		if req.OffsetID != 7 {
			t.Fatal("story offset", req.OffsetID)
		}
		*out.(*tg.StoriesStories) = tg.StoriesStories{Count: 2, Stories: []tg.StoryItemClass{&tg.StoryItem{ID: 6, Date: 10, Caption: "Story", Media: &tg.MessageMediaPhoto{Photo: &tg.Photo{ID: 100, AccessHash: 1, FileReference: []byte{1}, Sizes: []tg.PhotoSizeClass{&tg.PhotoSize{Type: "x", W: 800, H: 600, Size: 100}}}}}}}
		return nil
	}))
	p, e := s.SharedCollection(context.Background(), 42, model.SharedStories, "7", 1)
	if e != nil || len(p.Messages) != 1 || p.Next != "6" || p.Messages[0].Key.MessageID != -6 {
		t.Fatal(p, e)
	}
	history, e := s.Cache().Around(context.Background(), 42, 0, 100)
	if e != nil || len(history) != 0 {
		t.Fatal("story polluted message cache", history, e)
	}
	if _, ok := s.history.refs[p.Messages[0].Media.ID]; !ok {
		t.Fatal("story download missing")
	}
}
