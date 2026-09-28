// SPDX-License-Identifier: Unlicense OR MIT

package video

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestCustomFFmpegPlayer(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink fixture needs Unix")
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip(err)
	}
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "custom ffmpeg")
	if err := os.Symlink(ffmpeg, path); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(ffprobe, filepath.Join(dir, "ffprobe")); err != nil {
		t.Fatal(err)
	}
	media := filepath.Join(dir, "clip.mp4")
	if out, err := exec.Command(ffmpeg, "-v", "error", "-f", "lavfi", "-i", "color=c=red:s=16x16:d=0.2", "-c:v", "mpeg4", media).CombinedOutput(); err != nil {
		t.Fatalf("fixture: %v %s", err, out)
	}
	t.Setenv("PATH", dir)
	if ResolveFFmpeg("") != "" {
		t.Fatal("unexpected system FFmpeg")
	}
	if got := ResolveFFmpeg(path); got != path {
		t.Fatalf("custom path: %q", got)
	}
	if _, err := CheckFFmpeg(context.Background(), path); err != nil {
		t.Fatal(err)
	}
	p, err := NewPlayerWithFFmpeg(media, 32, path)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Stop()
	p.Start()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		im, err := p.Frame()
		if err != nil {
			t.Fatal(err)
		}
		if im != nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("custom FFmpeg did not produce a frame")
}
