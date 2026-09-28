// SPDX-License-Identifier: Unlicense OR MIT

//go:build !windows

package player_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"komarugram/pkg/player"
)

// script writes an executable shell script printing banner for any
// arguments.
func script(t *testing.T, banner string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nprintf '%s\\n' '"+banner+"'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestCheck feeds Check the players installed here, each other's paths and
// programs that are no player, and checks it tells them apart.
func TestCheck(t *testing.T) {
	ctx := context.Background()
	for _, kind := range player.Kinds {
		for _, path := range playerPaths(kind) {
			got, err := player.Check(ctx, kind, path)
			if err != nil {
				t.Errorf("%s as %s: %v", path, kind, err)
			} else {
				t.Logf("%s: %s", path, got)
			}
			other := player.VLC
			if kind == player.VLC {
				other = player.MPV
			}
			if _, err := player.Check(ctx, other, path); !errors.Is(err, player.ErrWrongProgram) {
				t.Errorf("%s as %s: got %v, want ErrWrongProgram", path, other, err)
			}
		}
	}
	if ls, err := exec.LookPath("ls"); err == nil {
		if _, err := player.Check(ctx, player.VLC, ls); !errors.Is(err, player.ErrWrongProgram) {
			t.Errorf("ls as vlc: got %v, want ErrWrongProgram", err)
		}
	}
	var old *player.VersionError
	if _, err := player.Check(ctx, player.VLC, script(t, "VLC version 2.2.8 Weatherwax")); !errors.As(err, &old) {
		t.Errorf("VLC 2: got %v, want a VersionError", err)
	}
	if got, err := player.Check(ctx, player.VLC, script(t, "VLC version 3.0.21 Vetinari")); err != nil || got != "VLC 3.0.21" {
		t.Errorf("VLC 3 script: got %q, %v", got, err)
	}

	plain := filepath.Join(t.TempDir(), "video.mp4")
	if err := os.WriteFile(plain, []byte("not a program"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{plain, t.TempDir(), "vlc", filepath.Join(t.TempDir(), "missing")} {
		if _, err := player.Check(ctx, player.VLC, path); !errors.Is(err, player.ErrNotExecutable) {
			t.Errorf("%s: got %v, want ErrNotExecutable", path, err)
		}
	}
}

// TestOpenChecks makes sure Open does not run a file that is not the player,
// such as one named in a settings file edited by hand.
func TestOpenChecks(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "ran")
	fake := filepath.Join(t.TempDir(), "vlc")
	body := "#!/bin/sh\nfor a; do [ \"$a\" = --version ] && { echo 'not a player 1.0'; exit 0; }; done\ntouch " + marker + "\n"
	if err := os.WriteFile(fake, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := player.Open(context.Background(), player.VLC, fake, "video.mp4"); !errors.Is(err, player.ErrWrongProgram) {
		t.Errorf("got %v, want ErrWrongProgram", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Error("the file was run as a player")
	}
}
