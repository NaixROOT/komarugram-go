// SPDX-License-Identifier: Unlicense OR MIT
package model

import "context"

// ChatTheme uses the portable Telegram themeSettings and wallPaperSettings
// semantics. RGB colors have no alpha; desktop palette colors are RGBA.
type ChatTheme struct {
	ID, Title   string
	Light, Dark *ChatThemeStyle
}
type ChatThemeStyle struct {
	Dark                    bool
	Accent, OutAccent       uint32
	Incoming, Text, OutText uint32
	Outgoing                []uint32
	Animated                bool
	Wallpaper               *ChatWallpaper
}
type ChatWallpaper struct {
	Colors                []uint32
	Rotation, Intensity   int
	Blur, Motion, Pattern bool
	Media                 *MessageMedia
	// Image holds imported wallpaper bytes; network wallpapers use Media.
	Image []byte `json:",omitempty"`
	Tile  bool
}
type ChatAppearance struct {
	Theme     ChatTheme
	Wallpaper *ChatWallpaper
}
type ChatThemeSource interface {
	ChatAppearance(context.Context, int64) (ChatAppearance, error)
	ChatThemes(context.Context) ([]ChatTheme, error)
	SetChatTheme(context.Context, int64, string) error
}

// LocalChatThemeSource persists an appearance override in the account cache.
type LocalChatThemeSource interface {
	SetLocalChatTheme(context.Context, int64, *ChatTheme) error
}
