// SPDX-License-Identifier: Unlicense OR MIT

package chattheme

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"komarugram/internal/messenger/model"
)

var paletteEntry = regexp.MustCompile(`(?m)([A-Za-z_][A-Za-z_0-9]*)\s*:\s*(#[0-9a-fA-F]{6}(?:[0-9a-fA-F]{2})?|[A-Za-z_][A-Za-z_0-9]*)\s*;`)
var comments = regexp.MustCompile(`(?s)/\*.*?\*/|//[^\r\n]*`)

// ImportDesktop reads a palette or ZIP theme, without extracting paths.
// Only documented chat palette entries and the bundled wallpaper are used.
func ImportDesktop(data []byte) (model.ChatTheme, error) {
	if len(data) > MaxBytes {
		return model.ChatTheme{}, errors.New("theme exceeds 16 MiB")
	}
	palette := data
	var background []byte
	tile := false
	if bytes.HasPrefix(data, []byte("PK")) {
		z, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return model.ChatTheme{}, err
		}
		if len(z.File) > 128 {
			return model.ChatTheme{}, errors.New("too many theme archive entries")
		}
		files := map[string]*zip.File{}
		for _, f := range z.File {
			files[strings.ToLower(f.Name)] = f
		}
		read := func(names ...string) ([]byte, string, error) {
			for _, name := range names {
				if f := files[name]; f != nil {
					if f.UncompressedSize64 > MaxBytes {
						return nil, "", errors.New("theme entry exceeds 16 MiB")
					}
					r, err := f.Open()
					if err != nil {
						return nil, "", err
					}
					b, err := io.ReadAll(io.LimitReader(r, MaxBytes+1))
					r.Close()
					if err != nil {
						return nil, "", err
					}
					if len(b) > MaxBytes {
						return nil, "", errors.New("theme entry exceeds 16 MiB")
					}
					return b, name, nil
				}
			}
			return nil, "", nil
		}
		palette, _, err = read("colors.tdesktop-theme", "colors.tdesktop-palette")
		if err != nil {
			return model.ChatTheme{}, err
		}
		if palette == nil {
			return model.ChatTheme{}, errors.New("theme palette is missing")
		}
		var name string
		background, name, err = read("background.jpg", "background.png", "tiled.jpg", "tiled.png")
		if err != nil {
			return model.ChatTheme{}, err
		}
		tile = strings.HasPrefix(name, "tiled.")
		if background != nil {
			if _, err := Decode(background); err != nil {
				return model.ChatTheme{}, fmt.Errorf("theme wallpaper: %w", err)
			}
		}
	}
	entries := map[string]string{}
	for _, m := range paletteEntry.FindAllStringSubmatch(comments.ReplaceAllString(string(palette), ""), -1) {
		entries[m[1]] = m[2]
	}
	if len(entries) == 0 {
		return model.ChatTheme{}, errors.New("theme contains no palette entries")
	}
	resolved := map[string]uint32{}
	visiting := map[string]bool{}
	var resolve func(string) (uint32, error)
	resolve = func(key string) (uint32, error) {
		if v, ok := resolved[key]; ok {
			return v, nil
		}
		if visiting[key] {
			return 0, fmt.Errorf("cyclic palette alias: %s", key)
		}
		value, ok := entries[key]
		if !ok {
			return 0, fmt.Errorf("unknown palette alias: %s", key)
		}
		visiting[key] = true
		defer delete(visiting, key)
		var v uint32
		if strings.HasPrefix(value, "#") {
			n, e := strconv.ParseUint(value[1:], 16, 32)
			if e != nil {
				return 0, e
			}
			v = uint32(n)
			if len(value) == 7 {
				v = v<<8 | 255
			}
		} else {
			var e error
			v, e = resolve(value)
			if e != nil {
				return 0, e
			}
		}
		resolved[key] = v
		return v, nil
	}
	// Validate aliases even in palette entries outside the chat subset.
	for k := range entries {
		if _, err := resolve(k); err != nil {
			return model.ChatTheme{}, err
		}
	}
	get := func(k string, fallback uint32) uint32 {
		if v, ok := resolved[k]; ok {
			return v
		}
		return fallback
	}
	s := &model.ChatThemeStyle{Incoming: get("msgInBg", 0xffffffff), Text: get("historyTextInFg", 0x18212aff), OutText: get("historyTextOutFg", 0x18212aff), Accent: get("historyLinkInFg", 0x168acdff) >> 8, OutAccent: get("historyLinkOutFg", 0x168acdff) >> 8, Outgoing: []uint32{get("msgOutBg", 0xe1ffc7ff) >> 8}, Wallpaper: &model.ChatWallpaper{Colors: []uint32{get("historyBg", 0xd8e4dcff) >> 8}, Image: background, Tile: tile}}
	c := RGB(s.Incoming >> 8)
	s.Dark = int(c.R)+int(c.G)+int(c.B) < 384
	return model.ChatTheme{Title: "Desktop theme", Light: s, Dark: s}, nil
}
