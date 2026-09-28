// SPDX-License-Identifier: Unlicense OR MIT

package video

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"komarugram/pkg/program"
)

var (
	ErrNotExecutable = errors.New("FFmpeg path is not executable")
	ErrWrongProgram  = errors.New("the program is not FFmpeg")
)

// ResolveFFmpeg prefers a custom executable, falling back to PATH.
func ResolveFFmpeg(custom string) string {
	if filepath.IsAbs(custom) && program.IsExecutable(custom) {
		return custom
	}
	path, _ := exec.LookPath("ffmpeg")
	return path
}

// CheckFFmpeg validates a program explicitly selected by the user.
func CheckFFmpeg(ctx context.Context, path string) (string, error) {
	if !filepath.IsAbs(path) || !program.IsExecutable(path) {
		return "", ErrNotExecutable
	}
	banner, err := program.Banner(ctx, path, "-version")
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(banner, "ffmpeg version ") {
		return "", ErrWrongProgram
	}
	return banner, nil
}

func probeExecutable(ffmpeg string) string {
	if ffmpeg != "" {
		name := "ffprobe"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		sibling := filepath.Join(filepath.Dir(ResolveFFmpeg(ffmpeg)), name)
		if program.IsExecutable(sibling) {
			return sibling
		}
	}
	return "ffprobe"
}
