// SPDX-License-Identifier: Unlicense OR MIT

package chattheme

import (
	"math"

	"komarugram/internal/messenger/model"
)

// Preset names a built-in look of every chat, as Telegram Desktop's
// embedded themes: Classic and Day are light, Tinted and Night dark. The
// empty preset is the application's own colors, with no wallpaper.
type Preset string

const (
	PresetApp     Preset = ""
	PresetClassic Preset = "classic"
	PresetDay     Preset = "day"
	PresetTinted  Preset = "tinted"
	PresetNight   Preset = "night"
)

// Presets are the presets in the order they are offered.
var Presets = []Preset{PresetApp, PresetClassic, PresetDay, PresetTinted, PresetNight}

// Dark reports whether the preset is for a dark theme.
func (p Preset) Dark() bool { return p == PresetTinted || p == PresetNight }

// Known reports whether p is one of the presets.
func (p Preset) Known() bool {
	for _, q := range Presets {
		if p == q {
			return true
		}
	}
	return false
}

type presetData struct {
	style model.ChatThemeStyle
	// accent is the color the preset's accents are, which a chosen accent
	// replaces; accents are the ones offered.
	accent  uint32
	accents []uint32
}

// The chat colors of Telegram Desktop's embedded themes, from their
// palettes: msgInBg, msgOutBg, historyText*Fg, historyLink*Fg, msg*DateFg,
// and the background each theme brings.
var presets = map[Preset]presetData{
	PresetClassic: {
		style: model.ChatThemeStyle{Incoming: 0xffffffff, Text: 0x000000ff, OutText: 0x000000ff, Outgoing: []uint32{0xeffdde}, Accent: 0x168acd, OutAccent: 0x168acd, InDate: 0xa0acb6ff, OutDate: 0x6db566ff,
			Wallpaper: &model.ChatWallpaper{Colors: []uint32{0xdbddbb, 0x6ba587, 0xd5d88d, 0x88b884}, Intensity: 50, Pattern: true}},
		accent:  0x40a7e3,
		accents: []uint32{0x45bce7, 0x52b440, 0xd46c99, 0xdf8a49, 0x9978c8, 0xc55245, 0x687b98, 0xdea922},
	},
	PresetDay: {
		style: model.ChatThemeStyle{Incoming: 0xffffffff, Text: 0x000000ff, OutText: 0x000000ff, Outgoing: []uint32{0xdef1fd}, Accent: 0x168acd, OutAccent: 0x168acd, InDate: 0xa0acb6ff, OutDate: 0x86a8c2ff,
			Wallpaper: &model.ChatWallpaper{Colors: []uint32{0x74b4e0}}},
		accent:  0x40a7e3,
		accents: []uint32{0x45bce7, 0x52b440, 0xd46c99, 0xdf8a49, 0x9978c8, 0xc55245, 0x687b98, 0xdea922},
	},
	PresetTinted: {
		style: model.ChatThemeStyle{Dark: true, Incoming: 0x182533ff, Text: 0xf5f5f5ff, OutText: 0xe4ecf2ff, Outgoing: []uint32{0x2b5278}, Accent: 0x70baf5, OutAccent: 0x83caff, InDate: 0x6d7f8fff, OutDate: 0x7da8d3ff,
			Wallpaper: &model.ChatWallpaper{Colors: []uint32{0x0e1621}, Dark: true}},
		accent:  0x5288c1,
		accents: []uint32{0x58bfe8, 0x466f42, 0xaa6084, 0xa46d3c, 0x917bbd, 0xab5149, 0x697b97, 0x9b834b},
	},
	PresetNight: {
		style: model.ChatThemeStyle{Dark: true, Incoming: 0x33393fff, Text: 0xf5f5f5ff, OutText: 0xe4ecf2ff, Outgoing: []uint32{0x2a2f33}, Accent: 0x37e1cb, OutAccent: 0x37e1cb, InDate: 0x828d94ff, OutDate: 0x737f87ff,
			Wallpaper: &model.ChatWallpaper{Colors: []uint32{0x18191d}, Dark: true}},
		accent:  0x3fc1b0,
		accents: []uint32{0x60a8e7, 0x4e9c57, 0xca7896, 0xcc925c, 0xa58ed2, 0xd27570, 0x7b8799, 0xcbac67},
	},
}

// Accents are the accent colors offered for p, as RGB; the first is the
// preset's own. The application's colors have none.
func Accents(p Preset) []uint32 {
	d, ok := presets[p]
	if !ok {
		return nil
	}
	return append([]uint32{d.accent}, d.accents[1:]...)
}

// PresetAccent is the preset's own accent color.
func PresetAccent(p Preset) uint32 { return presets[p].accent }

// Style is the preset's style, its accent turned to accent unless that is
// zero, with its own wallpaper. It is nil for the application's colors.
func Style(p Preset, accent uint32) *model.ChatThemeStyle {
	d, ok := presets[p]
	if !ok {
		return nil
	}
	s := d.style
	s.Outgoing = append([]uint32(nil), s.Outgoing...)
	w := *s.Wallpaper
	w.Colors = append([]uint32(nil), w.Colors...)
	if w.Pattern {
		w.Image = DoodlePattern()
	}
	s.Wallpaper = &w
	if accent != 0 && accent&0xffffff != d.accent {
		Colorize(&s, d.accent, accent&0xffffff)
	}
	return &s
}

// Colorize turns the colors of s near the hue of from to the hue of to, as
// Telegram Desktop colors a theme by an accent: grays, as white bubbles
// and text, keep their colors, as do colors far from the accent's hue.
func Colorize(s *model.ChatThemeStyle, from, to uint32) {
	was, now := hsvOf(from), hsvOf(to)
	shift := func(rgb uint32) uint32 {
		c := hsvOf(rgb)
		d := math.Abs(c.h - was.h)
		d = math.Min(d, 360-d)
		// Grays have no hue to move.
		if d > 15 || c.s < .02 {
			return rgb
		}
		c.h = math.Mod(c.h+now.h-was.h+360, 360)
		if was.s > 0 {
			c.s = math.Min(1, c.s*now.s/was.s)
		}
		if was.v > 0 {
			c.v = math.Min(1, c.v*now.v/was.v)
		}
		return c.rgb()
	}
	rgba := func(v uint32) uint32 {
		if v == 0 {
			return 0
		}
		return shift(v>>8)<<8 | v&0xff
	}
	for i, v := range s.Outgoing {
		s.Outgoing[i] = shift(v)
	}
	s.Incoming = rgba(s.Incoming)
	s.Accent = shift(s.Accent)
	s.OutAccent = shift(s.OutAccent)
	s.InDate = rgba(s.InDate)
	s.OutDate = rgba(s.OutDate)
}

// ThemeStyle is the style of a Telegram chat theme's variant: the preset of
// its base theme, turned to its accent, with its message colors. Arctic,
// which Telegram Desktop has not, draws as Day.
func ThemeStyle(base Preset, accent, outAccent uint32, messages []uint32, animated bool) *model.ChatThemeStyle {
	s := Style(base, accent)
	if s == nil {
		s = Style(PresetDay, accent)
	}
	s.Wallpaper = nil
	s.Animated = animated
	if outAccent != 0 {
		s.OutAccent = outAccent
	}
	if len(messages) > 0 {
		s.Outgoing = append([]uint32(nil), messages...)
		// The text keeps its contrast on the bubbles' own colors.
		if light := luminance(messages); light < .55 {
			s.OutText, s.OutDate = 0xffffffff, 0xffffffb3
			if s.OutAccent == s.Accent || luminanceOf(s.OutAccent) < .6 {
				s.OutAccent = 0xffffff
			}
		} else {
			s.OutText = 0x000000ff
			s.OutDate = shadeOf(messages[0], .55)
		}
	}
	return s
}

// luminance is the average relative lightness of colors, from 0 to 1.
func luminance(colors []uint32) float64 {
	sum := 0.
	for _, v := range colors {
		sum += luminanceOf(v)
	}
	return sum / float64(len(colors))
}

func luminanceOf(v uint32) float64 {
	c := RGB(v)
	return (float64(c.R)*.299 + float64(c.G)*.587 + float64(c.B)*.114) / 255
}

// shadeOf is an opaque RGBA color of v's hue, darker by k.
func shadeOf(v uint32, k float64) uint32 {
	c := hsvOf(v)
	c.s = math.Min(1, c.s+.25)
	c.v *= k
	return c.rgb()<<8 | 0xff
}

type hsv struct{ h, s, v float64 }

func hsvOf(v uint32) hsv {
	c := RGB(v)
	r, g, b := float64(c.R)/255, float64(c.G)/255, float64(c.B)/255
	hi, lo := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	out := hsv{v: hi}
	if hi > 0 {
		out.s = (hi - lo) / hi
	}
	d := hi - lo
	switch {
	case d == 0:
	case hi == r:
		out.h = 60 * math.Mod((g-b)/d+6, 6)
	case hi == g:
		out.h = 60 * ((b-r)/d + 2)
	default:
		out.h = 60 * ((r-g)/d + 4)
	}
	return out
}

func (c hsv) rgb() uint32 {
	k := func(n float64) float64 {
		k := math.Mod(n+c.h/60, 6)
		return c.v - c.v*c.s*math.Max(0, math.Min(k, math.Min(4-k, 1)))
	}
	to := func(x float64) uint32 { return uint32(math.Round(math.Max(0, math.Min(1, x)) * 255)) }
	return to(k(5))<<16 | to(k(3))<<8 | to(k(1))
}
