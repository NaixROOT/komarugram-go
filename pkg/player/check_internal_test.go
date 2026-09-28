// SPDX-License-Identifier: Unlicense OR MIT

package player

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestParseBanner(t *testing.T) {
	for _, c := range []struct {
		kind    Kind
		banner  string
		want    string
		wrong   bool
		version bool
	}{
		{kind: MPV, banner: "mpv 0.37.0 Copyright © 2000-2023 mpv/MPlayer/mplayer2 projects", want: "mpv 0.37.0"},
		{kind: MPV, banner: "mpv v0.38.0-617-g7e7ed4d0 Copyright © 2000-2024 mpv/MPlayer/mplayer2 projects", want: "mpv 0.38.0"},
		{kind: MPV, banner: "mpv 0.16.0 (C) 2000-2016 mpv/MPlayer/mplayer2 projects", version: true},
		{kind: VLC, banner: "VLC version 3.0.20 Vetinari (3.0.20-0-g6f0d0ab126b)", want: "VLC 3.0.20"},
		{kind: VLC, banner: "VLC media player version 3.0.21 Vetinari", want: "VLC 3.0.21"},
		{kind: VLC, banner: "VLC version 2.2.8 Weatherwax (2.2.8-0-g1b8e4a6)", version: true},
		{kind: VLC, banner: "VLC version 4.0.0-dev Otto Chriek", version: true},
		// The other player, and programs that are no player at all.
		{kind: VLC, banner: "mpv 0.37.0 Copyright © 2000-2023", wrong: true},
		{kind: MPV, banner: "rm (GNU coreutils) 9.4", wrong: true},
		{kind: VLC, banner: "Версия VLC 3.0.20 Vetinari", wrong: true},
	} {
		got, err := parseBanner(c.kind, c.banner)
		var versionErr *VersionError
		switch {
		case c.wrong && !errors.Is(err, ErrWrongProgram):
			t.Errorf("%q as %s: got %q, %v; want ErrWrongProgram", c.banner, c.kind, got, err)
		case c.version && !errors.As(err, &versionErr):
			t.Errorf("%q as %s: got %q, %v; want a VersionError", c.banner, c.kind, got, err)
		case !c.wrong && !c.version && (err != nil || got != c.want):
			t.Errorf("%q as %s: got %q, %v; want %q", c.banner, c.kind, got, err, c.want)
		}
	}
}

func TestIsSnap(t *testing.T) {
	dir := t.TempDir()
	launcher := filepath.Join(dir, "snap")
	if err := os.WriteFile(launcher, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "vlc")
	if err := os.Symlink(launcher, link); err != nil {
		t.Fatal(err)
	}
	for path, want := range map[string]bool{
		"/snap/bin/vlc":             true,
		"/snap/vlc/current/bin/vlc": true,
		link:                        true,
		"/usr/bin/vlc":              false,
	} {
		if got := isSnap(path); got != want {
			t.Errorf("isSnap(%q) = %t, want %t", path, got, want)
		}
	}
}
