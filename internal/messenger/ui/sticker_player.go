// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"

	"gio-mw/widget/radio"

	"gioui.org/layout"

	"komarugram/internal/messenger/localization"
	"komarugram/pkg/video"
)

type stickerPlayerSettings struct {
	program *programSetting
	radios  *radio.Radios[string]
	chosen  func() string
	choose  func(string)
}

func newStickerPlayerSettings() *stickerPlayerSettings {
	s := &stickerPlayerSettings{}
	s.program = &programSetting{
		title:  "FFmpeg",
		custom: func() string { return "" }, save: func(string) {},
		check: video.CheckFFmpeg,
		found: func(ctx context.Context) (string, string, error) {
			path := video.ResolveFFmpeg("")
			if path == "" {
				return "", "", nil
			}
			about, err := video.CheckFFmpeg(ctx, path)
			return path, about, err
		},
	}
	s.radios = radio.NewRadios([]string{"ffmpeg", "wasm"}, "wasm", func(value string) {
		if s.choose != nil {
			s.choose(value)
		}
	})
	return s
}

func (s *stickerPlayerSettings) current() string {
	if s.chosen != nil && s.chosen() != "" {
		return s.chosen()
	}
	if video.ResolveFFmpeg(s.program.custom()) != "" {
		return "ffmpeg"
	}
	return "wasm"
}

func (s *stickerPlayerSettings) Update(gtx layout.Context) {
	s.program.Update(gtx)
	s.radios.SetValue(s.current())
	s.radios.Update(gtx)
}

func (s *stickerPlayerSettings) Layout(gtx layout.Context, l localization.Catalog) layout.Dimensions {
	return settingsChoiceCard(gtx, l.T("sticker_player.title"), "", func(gtx layout.Context) layout.Dimensions {
		return s.radios.Layout(gtx, radio.LeadingKind, map[string]string{"ffmpeg": "FFmpeg", "wasm": l.T("sticker_player.wasm")})
	})
}
