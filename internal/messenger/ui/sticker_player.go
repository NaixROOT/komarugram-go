// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"

	"gio-mw/widget/radio"

	"gioui.org/layout"

	"komarugram/internal/messenger/localization"
	"komarugram/pkg/video"
)

// decoderSettings are the internal players: one FFmpeg, and for video
// stickers (WebM) and for GIFs and animated avatars (MP4) the choice between
// it and a WASM sandbox.
type decoderSettings struct {
	program              *programSetting
	stickers, animations *decoderChoice
}

// decoderChoice is the player of one kind of media.
type decoderChoice struct {
	// title and hint are localization keys; hint may be empty.
	title, hint string
	program     *programSetting
	radios      *radio.Radios[string]
	chosen      func() string
	choose      func(string)
}

func newDecoderSettings() *decoderSettings {
	program := &programSetting{
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
	return &decoderSettings{
		program:    program,
		stickers:   newDecoderChoice(program, "sticker_player.title", ""),
		animations: newDecoderChoice(program, "sticker_player.mp4_title", "sticker_player.mp4_hint"),
	}
}

func newDecoderChoice(program *programSetting, title, hint string) *decoderChoice {
	c := &decoderChoice{title: title, hint: hint, program: program}
	c.radios = radio.NewRadios([]string{"ffmpeg", "wasm"}, "wasm", func(value string) {
		if c.choose != nil {
			c.choose(value)
		}
	})
	return c
}

func (c *decoderChoice) current() string {
	if c.chosen != nil && c.chosen() != "" {
		return c.chosen()
	}
	if video.ResolveFFmpeg(c.program.custom()) != "" {
		return "ffmpeg"
	}
	return "wasm"
}

func (s *decoderSettings) Update(gtx layout.Context) {
	s.program.Update(gtx)
	for _, c := range []*decoderChoice{s.stickers, s.animations} {
		c.radios.SetValue(c.current())
		c.radios.Update(gtx)
	}
}

func (c *decoderChoice) Layout(gtx layout.Context, l localization.Catalog) layout.Dimensions {
	hint := ""
	if c.hint != "" {
		hint = l.T(c.hint)
	}
	return settingsChoiceCard(gtx, l.T(c.title), hint, func(gtx layout.Context) layout.Dimensions {
		return c.radios.Layout(gtx, radio.LeadingKind, map[string]string{"ffmpeg": "FFmpeg", "wasm": l.T("sticker_player.wasm")})
	})
}
