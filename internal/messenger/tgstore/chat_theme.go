// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"komarugram/internal/messenger/model"
	"sync"

	"github.com/gotd/td/tg"
)

// Serializing catalogue loads avoids simultaneous requests from windows of
// the same account. No UI mutex is held across a network operation.
type themeCatalogue struct {
	mu     sync.Mutex
	themes []model.ChatTheme
}

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
		_, _ = cache.Get(ctx, "chat-theme-catalogue", &cached)
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
		theme := model.ChatTheme{ID: t.Emoticon, Title: t.Emoticon + " " + t.Title}
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
		if err := cache.Put(ctx, "chat-theme-catalogue", out); err != nil {
			return nil, err
		}
	}
	s.themes.themes = out
	return append([]model.ChatTheme(nil), out...), nil
}
func (s *Store) themeStyle(ctx context.Context, t tg.ThemeSettings) *model.ChatThemeStyle {
	style := &model.ChatThemeStyle{Accent: uint32(t.AccentColor) & 0xffffff, OutAccent: uint32(t.OutboxAccentColor) & 0xffffff, Animated: t.MessageColorsAnimated}
	switch t.BaseTheme.(type) {
	case *tg.BaseThemeNight, *tg.BaseThemeTinted:
		style.Dark = true
	}
	if style.Dark {
		style.Incoming = 0x182533ff
		style.Text = 0xffffffff
		style.OutText = 0xffffffff
	} else {
		style.Incoming = 0xffffffff
		style.Text = 0x18212aff
		style.OutText = 0x18212aff
	}
	if style.OutAccent == 0 {
		style.OutAccent = style.Accent
	}
	for _, v := range t.MessageColors {
		style.Outgoing = append(style.Outgoing, uint32(v)&0xffffff)
	}
	if len(style.Outgoing) == 0 {
		if style.Dark {
			style.Outgoing = []uint32{0x2b5278}
		} else {
			style.Outgoing = []uint32{0xe1ffc7}
		}
	}
	if len(style.Outgoing) > 0 {
		var light uint32
		for _, v := range style.Outgoing {
			light += (v>>16&255)*299 + (v>>8&255)*587 + (v&255)*114
		}
		if light/uint32(len(style.Outgoing)) < 135000 {
			style.OutText = 0xffffffff
		} else {
			style.OutText = 0x18212aff
		}
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
	case *tg.WallPaper:
		settings = p.Settings
		w.Pattern = p.Pattern
		if d, ok := p.Document.(*tg.Document); ok {
			_, media, ref := documentMedia(d)
			w.Media = media
			ref.WallpaperID, ref.WallpaperHash = p.ID, p.AccessHash
			c := s.history
			c.mu.Lock()
			c.refs[media.ID] = *ref
			cache := c.cache
			c.mu.Unlock()
			if cache != nil {
				_ = cache.Put(ctx, "ref/"+media.ID, ref)
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
func (s *Store) ChatAppearance(ctx context.Context, chat int64) (model.ChatAppearance, error) {
	result := model.ChatAppearance{}
	if cache := s.Cache(); cache != nil {
		var local *model.ChatTheme
		if ok, e := cache.Get(ctx, fmt.Sprintf("local-theme/%d", chat), &local); e == nil && ok && local != nil {
			return model.ChatAppearance{Theme: *local}, nil
		}
	}
	// The full peer is authoritative, including an explicitly selected wallpaper.
	opctx, api, peer, _, done, err := s.peerOperation(ctx, chat)
	if err != nil {
		if cache := s.Cache(); cache != nil {
			if ok, e := cache.Get(ctx, fmt.Sprintf("appearance/%d", chat), &result); ok && e == nil {
				return result, nil
			}
		}
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
		if err := cache.Put(ctx, fmt.Sprintf("appearance/%d", chat), result); err != nil {
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
func (s *Store) invalidateChatTheme(chat int64) {
	s.mu.Lock()
	if s.themeRevisions == nil {
		s.themeRevisions = map[int64]uint64{}
	}
	s.themeRevisions[chat]++
	s.mu.Unlock()
}
