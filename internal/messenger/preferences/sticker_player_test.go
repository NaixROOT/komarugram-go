// SPDX-License-Identifier: Unlicense OR MIT

package preferences

import (
	"path/filepath"
	"testing"
)

func TestStickerPlayerPreferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	s, err := OpenPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Global().StickerPlayer != "" {
		t.Fatal("default must select an available decoder automatically")
	}
	if err := s.SetFFmpegPath("/custom path/ffmpeg"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetStickerPlayer("wasm"); err != nil {
		t.Fatal(err)
	}
	s, err = OpenPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Global().FFmpegPath != "/custom path/ffmpeg" || s.Global().StickerPlayer != "wasm" {
		t.Fatal("sticker settings not persisted")
	}
	if err := s.SetStickerPlayer("unknown"); err == nil {
		t.Fatal("invalid player accepted")
	}
	if s.Global().StickerPlayer != "wasm" {
		t.Fatal("invalid player changed settings")
	}
}
