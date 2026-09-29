// SPDX-License-Identifier: Unlicense OR MIT

// Package preferences persists settings shared by all messenger windows.
// Account-specific settings deliberately live in a separate namespace in the
// file format so they can be added without changing the meaning of globals.
package preferences

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"gio-mw/exp/powersave"

	"komarugram/pkg/miniapp"
	"komarugram/pkg/player"
)

const version = 1

// Theme is the global color-theme preference.
type Theme int

const (
	ThemeAuto Theme = iota
	ThemeLight
	ThemeDark
)

// ComposerStyle is how the message composer sits in a chat.
type ComposerStyle int

const (
	// ComposerFloating is a rounded capsule floating over the history.
	ComposerFloating ComposerStyle = iota
	// ComposerClassic is a full-width bar below the history.
	ComposerClassic
)

// Global contains preferences which currently apply to every account and
// every window.
type Global struct {
	Theme          Theme           `json:"theme"`
	Language       string          `json:"language"`
	LastAccountID  string          `json:"last_account_id,omitempty"`
	MotionMode     powersave.Mode  `json:"motion_mode"`
	LowBattery     int             `json:"low_battery"`
	MiniAppStorage miniapp.Storage `json:"mini_app_storage"`
	// VisualPrivacy hides phone numbers from the interface and covers the
	// account's identifiers in its profile with spoilers, for showing the
	// screen to others: streams, recordings, screenshots.
	VisualPrivacy bool `json:"visual_privacy,omitempty"`
	// StreamerMode hides the windows from screen capture, as AyuGram's
	// Streamer Mode, where the platform allows it.
	StreamerMode bool `json:"streamer_mode,omitempty"`
	// LocalPremium makes the accounts signed in here look Premium to
	// themselves, as AyuGram's Local Premium: the star beside their names,
	// and the Premium section of the settings. Telegram does not know of it,
	// so nothing else changes.
	LocalPremium bool `json:"local_premium,omitempty"`
	// Window locking only covers the UI; account connections stay running.
	AutoLockMinutes int           `json:"auto_lock_minutes,omitempty"`
	LockOnMinimize  bool          `json:"lock_on_minimize,omitempty"`
	LockOnClose     bool          `json:"lock_on_close,omitempty"`
	Composer        ComposerStyle `json:"composer,omitempty"`
	// ComposerBlur blurs the history behind the floating composer, while
	// animations are on.
	ComposerBlur bool `json:"composer_blur"`
	// Player is the external player videos open in; empty until the user
	// chooses one, which is asked only when more than one is installed.
	Player player.Kind `json:"player,omitempty"`
	// MPVPath and VLCPath are the players the user pointed at, which
	// player.Check accepted; empty for the ones found on the system.
	MPVPath string `json:"mpv_path,omitempty"`
	VLCPath string `json:"vlc_path,omitempty"`
	// BrowserPath is the browser for Mini Apps the user pointed at, which
	// miniapp.CheckBrowser accepted; empty for the one found.
	BrowserPath string `json:"browser_path,omitempty"`
	FFmpegPath  string `json:"ffmpeg_path,omitempty"`
	// StickerPlayer is empty for automatic selection, or ffmpeg/wasm.
	StickerPlayer string `json:"sticker_player,omitempty"`
	// AnimationPlayer plays GIFs and animated avatars, which are MP4: empty
	// for automatic selection, or ffmpeg/wasm.
	AnimationPlayer string `json:"animation_player,omitempty"`
	// AudioPlayer plays voice messages and music: empty or wasm for the
	// client's own, external for mpv or VLC.
	AudioPlayer string `json:"audio_player,omitempty"`
	// Ghost is what the accounts tell others of themselves, as AyuGram's
	// Ghost Mode, the same for all of them.
	Ghost Ghost `json:"ghost"`
	// Keep is what the cache keeps that Telegram takes back, as AyuGram's
	// saved deleted messages and edits history.
	Keep Keep `json:"keep"`
	// Filters hide messages, as AyuGram's message filters.
	Filters Filters `json:"filters"`
	// Look is how messages and avatars are drawn, as AyuGram's
	// customization.
	Look Look `json:"look"`
	// ConfirmSticker and ConfirmGIF ask before a sticker or a GIF chosen
	// in the composer is sent, as AyuGram's confirmations.
	ConfirmSticker bool `json:"confirm_sticker,omitempty"`
	ConfirmGIF     bool `json:"confirm_gif,omitempty"`
}

// Look is how messages and avatars are drawn.
type Look struct {
	// BubbleRadius rounds the corners of bubbles, in dp, up to 16.
	BubbleRadius int `json:"bubble_radius"`
	// AvatarCorners rounds avatars, from 0, square, to AvatarRound, a
	// circle.
	AvatarCorners int `json:"avatar_corners"`
	// Seconds shows the seconds of a message's time.
	Seconds bool `json:"seconds,omitempty"`
	// EditedMark and DeletedMark take the place of the marks of edited and
	// deleted messages; empty for the default ones.
	EditedMark  string `json:"edited_mark,omitempty"`
	DeletedMark string `json:"deleted_mark,omitempty"`
}

// The bounds of Look, as AyuGram's.
const (
	BubbleRadiusMax = 16
	AvatarRound     = 23
)

// Filters hide others' messages that match a pattern, or that blocked
// users sent; all are off by default, as in AyuGram.
type Filters struct {
	Enabled bool `json:"enabled,omitempty"`
	// InChats applies the filters in groups and private chats too; without
	// it they apply in channels only.
	InChats bool `json:"in_chats,omitempty"`
	// HideBlocked hides what blocked users sent, in every chat.
	HideBlocked bool            `json:"hide_blocked,omitempty"`
	Patterns    []FilterPattern `json:"patterns,omitempty"`
}

// FilterPattern is a regular expression, in Go's syntax, that hides the
// messages it matches, or, Reversed, the ones it does not; in Chat alone,
// or in every chat for 0.
type FilterPattern struct {
	Text            string `json:"text"`
	Reversed        bool   `json:"reversed,omitempty"`
	CaseInsensitive bool   `json:"case_insensitive,omitempty"`
	Chat            int64  `json:"chat,omitempty"`
}

// Keep is what the cache keeps: see model.Keep. Both are on by default,
// as in AyuGram.
type Keep struct {
	Deleted bool `json:"deleted"`
	Edits   bool `json:"edits"`
}

// Ghost is what the accounts tell others: see model.Ghost. Nothing is told
// by default; a chat is read on sending to it or reacting in it.
type Ghost struct {
	SendRead       bool `json:"send_read,omitempty"`
	SendOnline     bool `json:"send_online,omitempty"`
	SendTyping     bool `json:"send_typing,omitempty"`
	ReadOnInteract bool `json:"read_on_interact"`
}

// Equal reports whether g and o are the same preferences, as saved.
func (g Global) Equal(o Global) bool {
	a, errA := json.Marshal(g)
	b, errB := json.Marshal(o)
	return errA == nil && errB == nil && string(a) == string(b)
}

// PlayerPaths are the players the user pointed at, by kind.
func (g Global) PlayerPaths() map[player.Kind]string {
	return map[player.Kind]string{player.MPV: g.MPVPath, player.VLC: g.VLCPath}
}

type fileData struct {
	Version int    `json:"version"`
	Global  Global `json:"global"`
	// Accounts reserves a stable namespace for preferences which will belong
	// to one Telegram account. There are no such preferences yet.
	Accounts map[string]json.RawMessage `json:"accounts"`
}

// Store is a process-wide, persistent and observable preference store.
type Store struct {
	mu          sync.Mutex
	path        string
	global      Global
	accounts    map[string]json.RawMessage
	subscribers map[uint64]func()
	nextID      uint64
}

func defaults() Global {
	return Global{
		Theme:          ThemeAuto,
		Language:       "ru",
		MotionMode:     powersave.ModeAuto,
		LowBattery:     powersave.DefaultLowBattery,
		MiniAppStorage: miniapp.Shared,
		ComposerBlur:   true,
		Ghost:          Ghost{ReadOnInteract: true},
		Keep:           Keep{Deleted: true, Edits: true},
		Look:           Look{BubbleRadius: BubbleRadiusMax, AvatarCorners: AvatarRound},
	}
}

// Open loads the process-wide settings from the application config directory.
func Open() (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return OpenPath(filepath.Join(dir, "komarugram-go", "settings.json"))
}

// OpenPath loads settings from path. It is also useful to isolate tests.
func OpenPath(path string) (*Store, error) {
	s := &Store{path: path, global: defaults(), accounts: make(map[string]json.RawMessage), subscribers: make(map[uint64]func())}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	data := fileData{Global: defaults()}
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	if data.Version != version {
		return nil, fmt.Errorf("unsupported settings version %d", data.Version)
	}
	if err := validate(data.Global); err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	s.global = data.Global
	if data.Accounts != nil {
		s.accounts = data.Accounts
	}
	return s, nil
}

// Memory returns a non-persistent store, primarily for demos and tests.
func Memory() *Store {
	return &Store{global: defaults(), accounts: make(map[string]json.RawMessage), subscribers: make(map[uint64]func())}
}

func validate(g Global) error {
	if g.Theme < ThemeAuto || g.Theme > ThemeDark {
		return errors.New("invalid theme")
	}
	if g.Language != "ru" && g.Language != "en" {
		return errors.New("invalid language")
	}
	if g.MotionMode < powersave.ModeAuto || g.MotionMode > powersave.ModeOff {
		return errors.New("invalid animation mode")
	}
	if g.LowBattery < 0 || g.LowBattery > 100 {
		return errors.New("invalid low-battery threshold")
	}
	if g.MiniAppStorage < miniapp.Ephemeral || g.MiniAppStorage > miniapp.Shared {
		return errors.New("invalid Mini App storage")
	}
	if g.Composer < ComposerFloating || g.Composer > ComposerClassic {
		return errors.New("invalid composer style")
	}
	if g.Look.BubbleRadius < 0 || g.Look.BubbleRadius > BubbleRadiusMax || g.Look.AvatarCorners < 0 || g.Look.AvatarCorners > AvatarRound {
		return errors.New("invalid look")
	}
	if g.StickerPlayer != "" && g.StickerPlayer != "ffmpeg" && g.StickerPlayer != "wasm" {
		return errors.New("invalid sticker player")
	}
	if g.AnimationPlayer != "" && g.AnimationPlayer != "ffmpeg" && g.AnimationPlayer != "wasm" {
		return errors.New("invalid animation player")
	}
	if g.AudioPlayer != "" && g.AudioPlayer != "wasm" && g.AudioPlayer != "external" {
		return errors.New("invalid audio player")
	}
	if g.Player != "" && g.Player != player.MPV && g.Player != player.VLC {
		return errors.New("invalid external player")
	}
	if g.AutoLockMinutes < 0 || g.AutoLockMinutes > 120 {
		return errors.New("invalid automatic lock delay")
	}
	return nil
}

// Global returns a consistent snapshot.
func (s *Store) Global() Global {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.global
}

func (s *Store) SetFFmpegPath(path string) error {
	return s.change(func(g *Global) { g.FFmpegPath = path })
}

func (s *Store) SetStickerPlayer(value string) error {
	return s.change(func(g *Global) { g.StickerPlayer = value })
}

func (s *Store) SetAnimationPlayer(value string) error {
	return s.change(func(g *Global) { g.AnimationPlayer = value })
}

func (s *Store) SetAudioPlayer(value string) error {
	return s.change(func(g *Global) { g.AudioPlayer = value })
}

func (s *Store) SetTheme(value Theme) error {
	return s.change(func(g *Global) { g.Theme = value })
}

// SetLanguage changes the UI language for all windows.
func (s *Store) SetLanguage(value string) error {
	return s.change(func(g *Global) { g.Language = value })
}

// SetLastAccount records which account should be restored on the next start.
// It is updated when an account window gains focus or closes.
func (s *Store) SetLastAccount(id string) error {
	return s.change(func(g *Global) { g.LastAccountID = id })
}

func (s *Store) SetMotion(mode powersave.Mode, lowBattery int) error {
	return s.change(func(g *Global) {
		g.MotionMode = mode
		g.LowBattery = lowBattery
	})
}

// SetVisualPrivacy switches visual privacy mode for every window.
func (s *Store) SetVisualPrivacy(on bool) error {
	return s.change(func(g *Global) { g.VisualPrivacy = on })
}

// SetLocalPremium makes the accounts look Premium to themselves, or not.
func (s *Store) SetLocalPremium(on bool) error {
	return s.change(func(g *Global) { g.LocalPremium = on })
}

// SetStreamerMode hides the windows from screen capture, or shows them.
func (s *Store) SetStreamerMode(on bool) error {
	return s.change(func(g *Global) { g.StreamerMode = on })
}

func (s *Store) SetWindowLock(minutes int, minimize, close bool) error {
	return s.change(func(g *Global) {
		g.AutoLockMinutes = minutes
		g.LockOnMinimize = minimize
		g.LockOnClose = close
	})
}

// SetGhost changes what the accounts tell others of themselves.
func (s *Store) SetGhost(g Ghost) error {
	return s.change(func(global *Global) { global.Ghost = g })
}

// SetKeep changes what the cache keeps that Telegram takes back.
func (s *Store) SetKeep(k Keep) error {
	return s.change(func(global *Global) { global.Keep = k })
}

// SetFilters changes the message filters.
func (s *Store) SetFilters(f Filters) error {
	f.Patterns = slices.Clone(f.Patterns)
	return s.change(func(global *Global) { global.Filters = f })
}

// SetLook changes how messages and avatars are drawn.
func (s *Store) SetLook(l Look) error {
	return s.change(func(global *Global) { global.Look = l })
}

// SetConfirmations chooses whether stickers and GIFs are sent only once
// confirmed.
func (s *Store) SetConfirmations(sticker, gif bool) error {
	return s.change(func(g *Global) { g.ConfirmSticker, g.ConfirmGIF = sticker, gif })
}

// SetComposer changes the message composer style for all windows.
func (s *Store) SetComposer(value ComposerStyle) error {
	return s.change(func(g *Global) { g.Composer = value })
}

// SetComposerBlur switches the blur behind the floating composer.
func (s *Store) SetComposerBlur(on bool) error {
	return s.change(func(g *Global) { g.ComposerBlur = on })
}

// SetPlayer chooses the external player for videos.
func (s *Store) SetPlayer(kind player.Kind) error {
	return s.change(func(g *Global) { g.Player = kind })
}

// SetPlayerPath points the player kind at path; "" goes back to the one
// found on the system. The caller checks the path with player.Check.
func (s *Store) SetPlayerPath(kind player.Kind, path string) error {
	return s.change(func(g *Global) {
		switch kind {
		case player.MPV:
			g.MPVPath = path
		case player.VLC:
			g.VLCPath = path
		}
	})
}

// SetBrowserPath points Mini Apps at the browser at path; "" goes back to the
// one found. The caller checks the path with miniapp.CheckBrowser.
func (s *Store) SetBrowserPath(path string) error {
	return s.change(func(g *Global) { g.BrowserPath = path })
}

func (s *Store) SetMiniAppStorage(value miniapp.Storage) error {
	return s.change(func(g *Global) { g.MiniAppStorage = value })
}

func (s *Store) change(update func(*Global)) error {
	s.mu.Lock()
	next := s.global
	update(&next)
	if next.Equal(s.global) {
		s.mu.Unlock()
		return nil
	}
	if err := validate(next); err != nil {
		s.mu.Unlock()
		return err
	}
	if err := s.writeLocked(next); err != nil {
		s.mu.Unlock()
		return err
	}
	s.global = next
	callbacks := make([]func(), 0, len(s.subscribers))
	for _, callback := range s.subscribers {
		callbacks = append(callbacks, callback)
	}
	s.mu.Unlock()
	for _, callback := range callbacks {
		callback()
	}
	return nil
}

func (s *Store) writeLocked(global Global) error {
	if s.path == "" {
		return nil
	}
	data, err := json.MarshalIndent(fileData{
		Version:  version,
		Global:   global,
		Accounts: s.accounts,
	}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o600); err != nil {
		return err
	}
	if err := os.Chmod(tmp, 0o600); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// Subscribe registers a callback invoked after a successful change. The
// returned function removes it.
func (s *Store) Subscribe(callback func()) func() {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	s.subscribers[id] = callback
	s.mu.Unlock()
	return func() {
		s.mu.Lock()
		delete(s.subscribers, id)
		s.mu.Unlock()
	}
}
