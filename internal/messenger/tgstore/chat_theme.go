// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"komarugram/internal/crash"
	"komarugram/internal/messenger/chattheme"
	"komarugram/internal/messenger/model"
	"strings"
	"sync"

	"github.com/gotd/td/tg"
)

// Serializing catalogue loads avoids simultaneous requests from windows of
// the same account. No UI mutex is held across a network operation.
type themeCatalogue struct {
	mu         sync.Mutex
	themes     []model.ChatTheme
	wallpapers []model.ChatWallpaper
}

// The cache keys of the catalogues; a new version is fetched again, as
// the styles cached under the old ones were worked out otherwise.
const (
	themeCatalogueKey     = "chat-theme-catalogue/2"
	wallpaperCatalogueKey = "wallpaper-catalogue"
)

func (s *Store) ChatThemes(ctx context.Context) ([]model.ChatTheme, error) {
	s.themes.mu.Lock()
	defer s.themes.mu.Unlock()
	if s.themes.themes != nil {
		return append([]model.ChatTheme(nil), s.themes.themes...), nil
	}
	c := s.history
	c.mu.Lock()
	api, cache := c.api, c.cache
	if c.closing {
		c.mu.Unlock()
		return nil, model.ErrProfileOffline
	}
	c.wg.Add(1)
	c.mu.Unlock()
	defer c.wg.Done()
	var cached []model.ChatTheme
	if cache != nil {
		_, _ = cache.Get(ctx, themeCatalogueKey, &cached)
	}
	if api == nil {
		if cached != nil {
			return cached, nil
		}
		return nil, model.ErrProfileOffline
	}
	res, err := api.AccountGetChatThemes(ctx, 0)
	if err != nil {
		if cached != nil {
			return cached, nil
		}
		return nil, err
	}
	themes, ok := res.(*tg.AccountThemes)
	if !ok {
		return cached, nil
	}
	out := make([]model.ChatTheme, 0, len(themes.Themes))
	for _, t := range themes.Themes {
		theme := model.ChatTheme{ID: t.Emoticon, Title: strings.TrimSpace(t.Emoticon + " " + t.Title)}
		for _, settings := range t.Settings {
			style := s.themeStyle(ctx, settings)
			if style.Dark {
				theme.Dark = style
			} else {
				theme.Light = style
			}
		}
		if theme.ID != "" {
			out = append(out, theme)
		}
	}
	if cache != nil {
		if err := cache.Put(ctx, themeCatalogueKey, out); err != nil {
			return nil, err
		}
	}
	s.themes.themes = out
	return append([]model.ChatTheme(nil), out...), nil
}

// ChatWallpapers lists the wallpapers Telegram offers and the account
// saved, as Telegram Desktop's gallery; offline, as last listed.
func (s *Store) ChatWallpapers(ctx context.Context) ([]model.ChatWallpaper, error) {
	s.themes.mu.Lock()
	defer s.themes.mu.Unlock()
	if s.themes.wallpapers != nil {
		return append([]model.ChatWallpaper(nil), s.themes.wallpapers...), nil
	}
	c := s.history
	c.mu.Lock()
	api, cache := c.api, c.cache
	if c.closing {
		c.mu.Unlock()
		return nil, model.ErrProfileOffline
	}
	c.wg.Add(1)
	c.mu.Unlock()
	defer c.wg.Done()
	var cached []model.ChatWallpaper
	if cache != nil {
		_, _ = cache.Get(ctx, wallpaperCatalogueKey, &cached)
	}
	if api == nil {
		if cached != nil {
			return cached, nil
		}
		return nil, model.ErrProfileOffline
	}
	res, err := api.AccountGetWallPapers(ctx, 0)
	if err != nil {
		if cached != nil {
			return cached, nil
		}
		return nil, err
	}
	papers, ok := res.(*tg.AccountWallPapers)
	if !ok {
		return cached, nil
	}
	out := make([]model.ChatWallpaper, 0, len(papers.Wallpapers))
	for _, raw := range papers.Wallpapers {
		// A wallpaper with neither a picture nor colors has nothing to show.
		if w := s.wallpaper(ctx, raw); w != nil && (w.Media != nil || len(w.Colors) > 0) {
			out = append(out, *w)
		}
	}
	if cache != nil {
		if err := cache.Put(ctx, wallpaperCatalogueKey, out); err != nil {
			return nil, err
		}
	}
	s.themes.wallpapers = out
	return append([]model.ChatWallpaper(nil), out...), nil
}

// themeStyle is a chat theme's variant: the preset of its base theme turned
// to its accent, as Telegram Desktop draws a cloud theme.
func (s *Store) themeStyle(ctx context.Context, t tg.ThemeSettings) *model.ChatThemeStyle {
	base := chattheme.PresetDay
	switch t.BaseTheme.(type) {
	case *tg.BaseThemeClassic:
		base = chattheme.PresetClassic
	case *tg.BaseThemeNight:
		base = chattheme.PresetNight
	case *tg.BaseThemeTinted:
		base = chattheme.PresetTinted
	}
	accent := uint32(t.AccentColor) & 0xffffff
	var messages []uint32
	for _, v := range t.MessageColors {
		messages = append(messages, uint32(v)&0xffffff)
	}
	outAccent, _ := t.GetOutboxAccentColor()
	style := chattheme.ThemeStyle(base, accent, uint32(outAccent)&0xffffff, messages, t.MessageColorsAnimated)
	style.Accent = accent
	if outAccent == 0 && len(messages) == 0 {
		style.OutAccent = accent
	}
	style.Wallpaper = s.wallpaper(ctx, t.Wallpaper)
	return style
}

func (s *Store) wallpaper(ctx context.Context, raw tg.WallPaperClass) *model.ChatWallpaper {
	if raw == nil {
		return nil
	}
	w := &model.ChatWallpaper{}
	var settings tg.WallPaperSettings
	switch p := raw.(type) {
	case *tg.WallPaperNoFile:
		settings = p.Settings
		w.ID, w.Dark = p.ID, p.Dark
	case *tg.WallPaper:
		settings = p.Settings
		w.ID, w.Dark = p.ID, p.Dark
		w.Pattern = p.Pattern
		if d, ok := p.Document.(*tg.Document); ok {
			_, media, ref := documentMedia(d)
			w.Media = media
			ref.WallpaperID, ref.WallpaperHash = p.ID, p.AccessHash
			refs := map[string]fileLocation{media.ID: *ref}
			// The gallery shows the thumbnail, fetched on its own.
			if thumb := media.Thumbnail; thumb != nil && !strings.HasSuffix(thumb.ID, "/inline") {
				tr := *ref
				tr.Thumb = thumb.ID[strings.LastIndexByte(thumb.ID, '/')+1:]
				refs[thumb.ID] = tr
			}
			c := s.history
			c.mu.Lock()
			for id, r := range refs {
				c.refs[id] = r
			}
			cache := c.cache
			c.mu.Unlock()
			if cache != nil {
				for id, r := range refs {
					_ = cache.Put(ctx, "ref/"+id, r)
				}
			}
		}
	}
	if v, ok := settings.GetBackgroundColor(); ok {
		w.Colors = append(w.Colors, uint32(v)&0xffffff)
	}
	if v, ok := settings.GetSecondBackgroundColor(); ok {
		w.Colors = append(w.Colors, uint32(v)&0xffffff)
	}
	if v, ok := settings.GetThirdBackgroundColor(); ok {
		w.Colors = append(w.Colors, uint32(v)&0xffffff)
	}
	if v, ok := settings.GetFourthBackgroundColor(); ok {
		w.Colors = append(w.Colors, uint32(v)&0xffffff)
	}
	w.Rotation, w.Intensity, w.Blur, w.Motion = settings.Rotation, settings.Intensity, settings.Blur, settings.Motion
	return w
}

// The states of a chat's appearance in the cache: not asked again since
// the store started, asked, or changed by an update and to be asked again.
const (
	appearanceUnchecked uint8 = iota
	appearanceChecked
	appearanceStale
)

func appearanceKey(chat int64) string { return fmt.Sprintf("appearance/2/%d", chat) }

// ChatAppearance is the theme set for the chat only here, else the
// chat's own theme and wallpaper. The cached one comes at once, and is
// asked for again in the background once a session, so that a chat opens
// without waiting for Telegram; an update that changed it is waited for.
func (s *Store) ChatAppearance(ctx context.Context, chat int64) (model.ChatAppearance, error) {
	cache := s.Cache()
	if cache != nil {
		var local *model.ChatTheme
		if ok, e := cache.Get(ctx, fmt.Sprintf("local-theme/%d", chat), &local); e == nil && ok && local != nil {
			return model.ChatAppearance{Theme: *local, Local: true}, nil
		}
	}
	var cached model.ChatAppearance
	have := false
	if cache != nil {
		have, _ = cache.Get(ctx, appearanceKey(chat), &cached)
	}
	s.mu.Lock()
	if s.appearances == nil {
		s.appearances = map[int64]uint8{}
	}
	state := s.appearances[chat]
	if state != appearanceStale {
		s.appearances[chat] = appearanceChecked
	}
	s.mu.Unlock()
	if have && state != appearanceStale {
		if state == appearanceUnchecked {
			go s.revalidateAppearance(chat, cached)
		}
		return cached, nil
	}
	a, err := s.fetchAppearance(ctx, chat)
	if err != nil {
		if have {
			return cached, nil
		}
		return a, err
	}
	s.mu.Lock()
	s.appearances[chat] = appearanceChecked
	s.mu.Unlock()
	return a, nil
}

// revalidateAppearance asks for a chat's appearance, and tells the windows
// when it is not the cached one any more.
func (s *Store) revalidateAppearance(chat int64, cached model.ChatAppearance) {
	defer crash.Recover("chat appearance", nil)
	a, err := s.fetchAppearance(context.Background(), chat)
	if err != nil {
		return
	}
	was, _ := json.Marshal(cached)
	now, _ := json.Marshal(a)
	if !bytes.Equal(was, now) {
		s.bumpChatTheme(chat)
		if s.changed != nil {
			s.changed()
		}
	}
}

// fetchAppearance asks Telegram for a chat's theme and wallpaper, and
// caches them.
func (s *Store) fetchAppearance(ctx context.Context, chat int64) (model.ChatAppearance, error) {
	result := model.ChatAppearance{}
	// The full peer is authoritative, including an explicitly selected wallpaper.
	opctx, api, peer, _, done, err := s.peerOperation(ctx, chat)
	if err != nil {
		return result, err
	}
	defer done()
	ctx = opctx
	var selected tg.ChatThemeClass
	var emoji string
	var wallpaper tg.WallPaperClass
	switch peer.Kind {
	case "user":
		full, e := api.UsersGetFullUser(ctx, &tg.InputUser{UserID: peer.ID, AccessHash: peer.Hash})
		if e != nil {
			return result, e
		}
		selected = full.FullUser.Theme
		wallpaper = full.FullUser.Wallpaper
	case "channel":
		full, e := api.ChannelsGetFullChannel(ctx, &tg.InputChannel{ChannelID: peer.ID, AccessHash: peer.Hash})
		if e != nil {
			return result, e
		}
		if c, ok := full.FullChat.(*tg.ChannelFull); ok {
			emoji = c.ThemeEmoticon
			wallpaper = c.Wallpaper
		}
	case "chat":
		full, e := api.MessagesGetFullChat(ctx, peer.ID)
		if e != nil {
			return result, e
		}
		if c, ok := full.FullChat.(*tg.ChatFull); ok {
			emoji = c.ThemeEmoticon
		}
	}
	switch t := selected.(type) {
	case *tg.ChatTheme:
		emoji = t.Emoticon
	case *tg.ChatThemeUniqueGift:
		if gift, ok := t.Gift.(*tg.StarGiftUnique); ok {
			result.Theme.Title = gift.Title
		}
		for _, settings := range t.ThemeSettings {
			style := s.themeStyle(ctx, settings)
			if style.Dark {
				result.Theme.Dark = style
			} else {
				result.Theme.Light = style
			}
		}
	}
	if emoji != "" {
		themes, e := s.ChatThemes(ctx)
		if e != nil {
			return result, e
		}
		for _, theme := range themes {
			if theme.ID == emoji {
				result.Theme = theme
				break
			}
		}
	}
	result.Wallpaper = s.wallpaper(ctx, wallpaper)
	if cache := s.Cache(); cache != nil {
		if err := cache.Put(ctx, appearanceKey(chat), result); err != nil {
			return result, err
		}
	}
	return result, nil
}
func (s *Store) SetChatTheme(ctx context.Context, chat int64, id string) error {
	ctx, api, peer, _, done, err := s.peerOperation(ctx, chat)
	if err != nil {
		return err
	}
	defer done()
	var theme tg.InputChatThemeClass = &tg.InputChatThemeEmpty{}
	if id != "" {
		theme = &tg.InputChatTheme{Emoticon: id}
	}
	updates, err := api.MessagesSetChatTheme(ctx, &tg.MessagesSetChatThemeRequest{Peer: peer.input(), Theme: theme})
	if err != nil {
		return err
	}
	if err := s.Handle(ctx, updates); err != nil {
		return err
	}
	return s.SetLocalChatTheme(ctx, chat, nil)
}

func (s *Store) SetLocalChatTheme(ctx context.Context, chat int64, theme *model.ChatTheme) error {
	c := s.history
	c.mu.Lock()
	cache := c.cache
	if c.closing || cache == nil {
		c.mu.Unlock()
		return model.ErrProfileOffline
	}
	c.wg.Add(1)
	c.mu.Unlock()
	defer c.wg.Done()
	return cache.Put(ctx, fmt.Sprintf("local-theme/%d", chat), theme)
}

func (s *Store) ChatThemeRevision(chat int64) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.themeRevisions[chat]
}

// invalidateChatTheme tells the windows that an update changed a chat's
// theme, which is asked for again.
func (s *Store) invalidateChatTheme(chat int64) {
	s.mu.Lock()
	if s.appearances == nil {
		s.appearances = map[int64]uint8{}
	}
	s.appearances[chat] = appearanceStale
	s.mu.Unlock()
	s.bumpChatTheme(chat)
}

func (s *Store) bumpChatTheme(chat int64) {
	s.mu.Lock()
	if s.themeRevisions == nil {
		s.themeRevisions = map[int64]uint64{}
	}
	s.themeRevisions[chat]++
	s.mu.Unlock()
}
