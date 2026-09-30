// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"os"
	"testing"

	"gioui.org/layout"

	"komarugram/internal/messenger/localization"
	"komarugram/pkg/player"
)

func TestPlayerFor(t *testing.T) {
	both := []player.Kind{player.MPV, player.VLC}
	for _, c := range []struct {
		name      string
		chosen    player.Kind
		installed []player.Kind
		kind      player.Kind
		ask       bool
	}{
		{"none installed", player.VLC, nil, "", false},
		{"only one", "", []player.Kind{player.VLC}, player.VLC, false},
		{"chosen one removed", player.MPV, []player.Kind{player.VLC}, player.VLC, false},
		{"both, not chosen", "", both, "", true},
		{"both, chosen", player.VLC, both, player.VLC, false},
		{"only the browser", "", []player.Kind{player.Chromium}, player.Chromium, false},
		{"a player before the browser", "", []player.Kind{player.VLC, player.Chromium}, player.VLC, false},
		{"both before the browser", "", []player.Kind{player.MPV, player.VLC, player.Chromium}, "", true},
		{"the browser, chosen", player.Chromium, []player.Kind{player.MPV, player.VLC, player.Chromium}, player.Chromium, false},
		{"the browser chosen, then gone", player.Chromium, []player.Kind{player.VLC}, player.VLC, false},
	} {
		kind, ask := playerFor(c.chosen, c.installed)
		if kind != c.kind || ask != c.ask {
			t.Errorf("%s: got %q, ask %t; want %q, ask %t", c.name, kind, ask, c.kind, c.ask)
		}
	}
}

// TestRenderPlayerChoice saves the dialog that asks which player to open
// videos in, for looking at it:
//
//	PLAYER_PNG=/tmp/player.png go test ./internal/messenger/ui -run RenderPlayerChoice
func TestRenderPlayerChoice(t *testing.T) {
	path := os.Getenv("PLAYER_PNG")
	if path == "" {
		t.Skip("set PLAYER_PNG to a file name")
	}
	p := &chatPage{}
	p.playerChoice.kinds = []player.Kind{player.MPV, player.VLC}
	p.playerChoice.modal.Open()
	renderFrames(t, image.Pt(700, 500), path, func(gtx layout.Context) {
		p.playerDialog(gtx, localization.For("ru"))
	})
}
