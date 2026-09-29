// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"os"
	"path/filepath"
	"testing"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"

	"gioui.org/layout"
)

// TestRenderAudio draws the demo's voice messages and music in the history:
// voice messages with the waveform Telegram sends and without, whose
// waveforms are worked out as they show (the M4A one's decoder is fetched),
// and music with its bar, paused at 60%: one plays at a time. In the light
// and the dark theme, into AUDIO_PNG_DIR:
//
//	AUDIO_PNG_DIR=/tmp/voice go test ./internal/messenger/ui -run RenderAudio
func TestRenderAudio(t *testing.T) {
	dir := os.Getenv("AUDIO_PNG_DIR")
	if dir == "" {
		t.Skip("set AUDIO_PNG_DIR to a directory")
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir("../../..")
	l := localization.For("ru")
	for _, dark := range []bool{false, true} {
		p, _, find := audioHarnessPage(t)
		p.images = &imageOps{}
		voiced := find("demo/voice")
		// Scroll so that the voice messages are at the top.
		draw := func(gtx layout.Context) {
			p.images.BeginFrame()
			p.media.BeginFrame()
			p.Layout(gtx, model.Chat{ID: 2}, l, false)
			p.media.EndFrame()
			p.images.EndFrame()
		}
		renderToast(t, filepath.Join(dir, "scratch.png"), image.Pt(600, 820), dark, draw)
		for i, m := range p.messages {
			if m.Key == voiced.Key {
				p.list.Position = layout.Position{First: i, BeforeEnd: true}
			}
		}
		music := find("demo/music")
		p.audio.toggle(p, music, 0.6)
		waitAudio(t, "the music did not start", func() bool { return p.audio.state(music).playing })
		p.audio.toggle(p, music, -1)
		theme := "light"
		if dark {
			theme = "dark"
		}
		renderToast(t, filepath.Join(dir, "audio-"+theme+".png"), image.Pt(600, 820), dark, draw)
	}
	os.Remove(filepath.Join(dir, "scratch.png"))
}
