// SPDX-License-Identifier: Unlicense OR MIT

// Package player drives an external video player: mpv or VLC.
//
// Playback inside the application window is handled by the video package; this
// is the other half, the one a messenger opens when the user wants the full
// player with a seek bar. The player runs as a separate process, so its
// decoders — and any bug in them — stay outside this one, and it is controlled
// over its IPC channel rather than by restarting it with new arguments.
package player

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"komarugram/pkg/program"
)

// Kind names an external player.
type Kind string

const (
	MPV Kind = "mpv"
	VLC Kind = "vlc"
)

// Kinds lists the supported players in the order they are preferred when
// the user has not chosen one.
var Kinds = []Kind{MPV, VLC}

// Title is the player's name as its authors write it.
func (k Kind) Title() string {
	if k == VLC {
		return "VLC"
	}
	return string(k)
}

// Status is what the player reports about the current file.
type Status struct {
	Running  bool
	Paused   bool
	Position time.Duration
	Duration time.Duration
	File     string
	Err      error
}

// Player is one running player process with its own window.
type Player interface {
	// Status reports the latest state the player sent.
	Status() Status
	// TogglePause flips playback without restarting the player.
	TogglePause() error
	// Seek moves by delta, forward or backward.
	Seek(delta time.Duration) error
	// Close stops the player.
	Close() error
}

// flatpakIDs are the flatpaks of the players.
var flatpakIDs = map[Kind]string{MPV: "io.mpv.Mpv", VLC: "org.videolan.VLC"}

// FlatpakID is the application id of the player's flatpak.
func (k Kind) FlatpakID() string { return flatpakIDs[k] }

// Find returns the executable of the player found on this system — on PATH,
// as a flatpak, or where its Windows installer puts it — or "" if there is
// none that can be driven here.
func (k Kind) Find() string {
	// mpv's IPC on Windows is a named pipe, which dial does not open yet:
	// offering mpv there would only fail after the timeout.
	if k == MPV && runtime.GOOS == "windows" {
		return ""
	}
	if path, err := exec.LookPath(string(k)); err == nil && !isSnap(path) {
		return path
	}
	if runtime.GOOS == "windows" {
		if k == VLC {
			// The installer does not put VLC on PATH.
			for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)", "ProgramW6432"} {
				if dir := os.Getenv(env); dir != "" {
					path := filepath.Join(dir, "VideoLAN", "VLC", "vlc.exe")
					if program.IsExecutable(path) {
						return path
					}
				}
			}
		}
		return ""
	}
	return program.FindFlatpak(flatpakIDs[k])
}

// Available reports whether the player was found on this system.
func (k Kind) Available() bool {
	return k.Find() != ""
}

// Resolve returns the executable to run for the player: custom, the path the
// user gave, while it is still there, and otherwise the one found.
func (k Kind) Resolve(custom string) string {
	if custom != "" && program.IsExecutable(custom) {
		return custom
	}
	return k.Find()
}

// Installed lists the players that can be started, in the order of Kinds;
// custom holds the paths the user gave.
func Installed(custom map[Kind]string) []Kind {
	var kinds []Kind
	for _, k := range Kinds {
		if k.Resolve(custom[k]) != "" {
			kinds = append(kinds, k)
		}
	}
	return kinds
}

// PrivateArgs are the arguments that keep a player from reading the user's
// scripts and extensions, going to the network on its own, and remembering
// what it played: the source is a loopback URL with a secret in it.
func (k Kind) PrivateArgs() []string {
	switch k {
	case MPV:
		return []string{"--no-config", "--load-scripts=no", "--ytdl=no"}
	case VLC:
		return []string{
			"--ignore-config",
			"--no-qt-recentplay",
			"--qt-continue=0",
			"--no-qt-privacy-ask",
			"--no-media-library",
			"--no-metadata-network-access",
		}
	}
	return nil
}

// Open starts the player at path — "" for the one found — on source, which
// may be a path or a URL, and waits for its IPC channel to come up. Extra
// arguments are appended to the player's command line.
func Open(ctx context.Context, kind Kind, path, source string, extra ...string) (Player, error) {
	if path == "" {
		path = kind.Find()
	}
	if path == "" {
		return nil, fmt.Errorf("%s is not installed", kind.Title())
	}
	// The path may come from a settings file edited by hand, or name a
	// file replaced since it was picked.
	if _, err := checked(ctx, kind, path); err != nil {
		return nil, err
	}
	switch kind {
	case MPV:
		return openMPV(ctx, path, source, extra)
	case VLC:
		return openVLC(ctx, path, source, extra)
	}
	return nil, fmt.Errorf("unknown player %q", kind)
}

// socketPath returns a fresh path for the Unix socket of the player at path,
// in a directory the player sees: a flatpak has a /tmp of its own. It is ""
// when there is no such directory.
func socketPath(kind Kind, path string) string {
	dir := os.TempDir()
	if id, ok := program.FlatpakApp(path); ok {
		dir = program.RuntimeDir(id)
	}
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, fmt.Sprintf("kitchen-%s-%d.sock", kind, time.Now().UnixNano()))
}

// dial waits for the player to open its IPC endpoint, which takes a moment
// after start.
func dial(ctx context.Context, network, address string) (net.Conn, error) {
	deadline := time.Now().Add(5 * time.Second)
	for {
		conn, err := net.Dial(network, address)
		if err == nil {
			return conn, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("player did not open %s: %w", address, err)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
}
