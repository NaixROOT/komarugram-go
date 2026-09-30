// SPDX-License-Identifier: Unlicense OR MIT
package model

import "context"

// ChatTheme uses the portable Telegram themeSettings and wallPaperSettings
// semantics. RGB colors have no alpha; desktop palette colors are RGBA.
type ChatTheme struct {
	ID, Title   string
	Light, Dark *ChatThemeStyle
}

// ChatThemeStyle is how the history of a chat is drawn in one of a theme's
// variants: the bubbles, their text and the wallpaper behind them.
type ChatThemeStyle struct {
	Dark                    bool
	Accent, OutAccent       uint32
	Incoming, Text, OutText uint32
	Outgoing                []uint32
	// InDate and OutDate are the RGBA colors of the time and the marks in
	// a message's footer; zero derives them from the text.
	InDate, OutDate uint32 `json:",omitempty"`
	Animated        bool
	Wallpaper       *ChatWallpaper
}

type ChatWallpaper struct {
	// ID is Telegram's wallpaper, zero for one made of colors or a file.
	ID                    int64 `json:",omitempty"`
	Colors                []uint32
	Rotation, Intensity   int
	Blur, Motion, Pattern bool
	// Dark is set on a wallpaper made for dark themes.
	Dark  bool `json:",omitempty"`
	Media *MessageMedia
	// Image holds imported wallpaper bytes; network wallpapers use Media.
	Image []byte `json:",omitempty"`
	Tile  bool
}

type ChatAppearance struct {
	Theme     ChatTheme
	Wallpaper *ChatWallpaper
	// Local is set when the theme was set for the chat only here.
	Local bool `json:",omitempty"`
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

// WallpaperSource lists the wallpapers to choose from for every chat, as
// Telegram's account.getWallPapers does. Their pictures are Media, which
// the store's Media downloads, with a Thumbnail for the gallery.
type WallpaperSource interface {
	ChatWallpapers(context.Context) ([]ChatWallpaper, error)
}
