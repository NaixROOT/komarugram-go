// SPDX-License-Identifier: Unlicense OR MIT

package mockstore

import (
	"context"
	"sync"

	"komarugram/internal/messenger/chattheme"
	"komarugram/internal/messenger/model"
)

// demoTheme is a chat theme as Telegram's emoji themes are made: a light
// and a dark variant over a pattern of the same four colors.
type demoTheme struct {
	emoji, title string
	colors       [4]uint32
	outgoing     []uint32
	accent       uint32
}

// The demo's chat themes, made up after the kinds Telegram offers: the
// real ones come from the server.
var demoThemes = []demoTheme{
	{"🐥", "Chick", [4]uint32{0xe6e3a0, 0xa9d67e, 0xf0dc8c, 0xbfdc86}, []uint32{0xf4fbc9}, 0x6f9f2e},
	{"⛄", "Snowman", [4]uint32{0xc3d8f0, 0x9fc4ec, 0xe2ecf8, 0xb3cff2}, []uint32{0xd8ecff}, 0x3a86d0},
	{"💎", "Diamond", [4]uint32{0xc5b6f6, 0x9fd2f3, 0xb6e8f6, 0xa7b8f5}, []uint32{0xd4e4ff, 0xe2d6ff}, 0x6a78e0},
	{"👨‍🏫", "Teacher", [4]uint32{0xa6d9b6, 0x7cc5c4, 0xcce6a8, 0x92cfb2}, []uint32{0xdcf6e0}, 0x2a9a76},
	{"🌷", "Tulip", [4]uint32{0xf4b6c7, 0xf6d7b2, 0xe8a6c8, 0xf7c6a8}, []uint32{0xffe2eb}, 0xd0567f},
	{"💜", "Heart", [4]uint32{0xc9a3ee, 0xefa7d2, 0xa89bf0, 0xf1b5dc}, []uint32{0xf1dbff, 0xe6c7ff}, 0x9656d4},
	{"🎄", "Fir", [4]uint32{0xf1a56b, 0xe7c55e, 0xe27b5a, 0xefb26d}, []uint32{0xfff0cb}, 0xca5b2c},
	{"🎮", "Game", [4]uint32{0x8fb3f2, 0xb88ef0, 0x80cff0, 0xc7a4f5}, []uint32{0xd0d8ff, 0xe1cfff}, 0x6465de},
}

var demoCatalogue = sync.OnceValue(func() []model.ChatTheme {
	pattern := chattheme.DoodlePattern()
	out := make([]model.ChatTheme, 0, len(demoThemes))
	for _, d := range demoThemes {
		light := chattheme.ThemeStyle(chattheme.PresetClassic, d.accent, 0, d.outgoing, false)
		light.Wallpaper = &model.ChatWallpaper{Colors: d.colors[:], Intensity: 50, Pattern: true, Image: pattern}
		// The dark variant shows the same colors through the pattern only,
		// as Telegram's dark patterns do.
		var dim []uint32
		for _, v := range d.outgoing {
			c := chattheme.RGB(v)
			dim = append(dim, uint32(c.R)*2/5<<16|uint32(c.G)*2/5<<8|uint32(c.B)*2/5)
		}
		dark := chattheme.ThemeStyle(chattheme.PresetTinted, lighten(d.accent), 0, dim, false)
		dark.Wallpaper = &model.ChatWallpaper{Colors: d.colors[:], Intensity: -50, Pattern: true, Dark: true, Image: pattern}
		out = append(out, model.ChatTheme{ID: d.emoji, Title: d.emoji + " " + d.title, Light: light, Dark: dark})
	}
	return out
})

// lighten brings an accent halfway to white, for dark bubbles.
func lighten(v uint32) uint32 {
	c := chattheme.RGB(v)
	return uint32(c.R/2+128)<<16 | uint32(c.G/2+128)<<8 | uint32(c.B/2+128)
}

// demoWallpapers are the demo's gallery: patterns, gradients, colors and
// photos, as Telegram's has.
var demoWallpapers = sync.OnceValue(func() []model.ChatWallpaper {
	pattern := chattheme.DoodlePattern()
	var out []model.ChatWallpaper
	id := int64(1)
	add := func(w model.ChatWallpaper) {
		w.ID = id
		id++
		out = append(out, w)
	}
	for _, d := range demoThemes[:6] {
		add(model.ChatWallpaper{Colors: d.colors[:], Intensity: 50, Pattern: true, Image: pattern})
	}
	for _, colors := range [][]uint32{{0xfec496, 0xdd6cb9, 0x962fbf, 0x4f5bd5}, {0x7fa381, 0xfff5c5, 0x336f55, 0xfbe37d}, {0xe4b2ea, 0x8376c2, 0xeab9d9, 0xb493e6}, {0xf7dd6d, 0x5ab5e0}} {
		add(model.ChatWallpaper{Colors: colors, Rotation: 45})
	}
	for _, c := range []uint32{0x74b4e0, 0xd8e4dc, 0xf5ead6, 0xb8c9d8} {
		add(model.ChatWallpaper{Colors: []uint32{c}})
	}
	for _, n := range []int{2, 8, 11} {
		m := demoPhoto(n)
		if len(m.Variants) > 0 {
			thumb := m.Variants[0]
			m.Thumbnail = &thumb
		}
		add(model.ChatWallpaper{Media: m})
	}
	for _, d := range []demoTheme{demoThemes[2], demoThemes[5]} {
		add(model.ChatWallpaper{Colors: d.colors[:], Intensity: -50, Pattern: true, Dark: true, Image: pattern})
	}
	return out
})

func (s *Store) ChatThemes(context.Context) ([]model.ChatTheme, error) {
	return append([]model.ChatTheme(nil), demoCatalogue()...), nil
}

func (s *Store) ChatWallpapers(context.Context) ([]model.ChatWallpaper, error) {
	return append([]model.ChatWallpaper(nil), demoWallpapers()...), nil
}

// ChatAppearance is the theme set here, else the one set in the chat.
func (s *Store) ChatAppearance(ctx context.Context, chat int64) (model.ChatAppearance, error) {
	s.mu.Lock()
	local, ok := s.themes[chat]
	emoji := s.chatThemes[chat]
	s.mu.Unlock()
	if ok {
		return model.ChatAppearance{Theme: local, Local: true}, nil
	}
	for _, t := range demoCatalogue() {
		if t.ID == emoji && emoji != "" {
			return model.ChatAppearance{Theme: t}, nil
		}
	}
	return model.ChatAppearance{}, nil
}

// SetChatTheme sets the theme for all in the chat, which the one set here
// no longer hides.
func (s *Store) SetChatTheme(ctx context.Context, chat int64, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.chatThemes == nil {
		s.chatThemes = map[int64]string{}
	}
	s.chatThemes[chat] = id
	delete(s.themes, chat)
	return nil
}

func (s *Store) SetLocalChatTheme(ctx context.Context, chat int64, theme *model.ChatTheme) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.themes == nil {
		s.themes = map[int64]model.ChatTheme{}
	}
	if theme == nil {
		delete(s.themes, chat)
	} else {
		s.themes[chat] = *theme
	}
	return nil
}
