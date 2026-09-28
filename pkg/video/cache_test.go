// SPDX-License-Identifier: Unlicense OR MIT

package video_test

import (
	"path/filepath"
	"testing"
	"time"

	"komarugram/pkg/video"
)

// TestDecodeClips reports what caching the clips in ./videos costs at a few
// tile sizes and frame rates. It is a measurement, not an assertion.
func TestDecodeClips(t *testing.T) {
	paths, err := filepath.Glob("../../videos/*.mp4")
	if err != nil || len(paths) == 0 {
		t.Skip("no ../../videos/*.mp4 to measure")
	}

	for _, cfg := range []struct {
		height    int
		fps       float64
		maxFrames int
		format    video.Format
	}{
		{240, 30, 4096, video.FormatRGBA},
		{240, 15, 4096, video.FormatRGBA},
		{180, 12, 4096, video.FormatRGBA},
		{240, 15, 90, video.FormatRGBA},
		{240, 15, 90, video.FormatYCbCr},
		{240, 15, 4096, video.FormatYCbCr},
	} {
		var totalBytes, totalFrames int
		var totalDecode time.Duration
		t.Logf("--- tile height %d, %g fps, cap %d frames, %v ---", cfg.height, cfg.fps, cfg.maxFrames, cfg.format)
		for _, path := range paths {
			start := time.Now()
			clip, err := video.DecodeClip(path, cfg.height, cfg.fps, cfg.maxFrames, cfg.format)
			if err != nil {
				t.Fatalf("%s: %v", path, err)
			}
			took := time.Since(start)
			totalBytes += clip.Bytes()
			totalFrames += clip.Len()
			totalDecode += took
			t.Logf("  %-12s %dx%d %4d frames %6.1f MB  decoded in %5.0f ms (%.1fx realtime)",
				filepath.Base(path), clip.Size.X, clip.Size.Y, clip.Len(),
				float64(clip.Bytes())/1e6, float64(took.Milliseconds()),
				clip.Duration().Seconds()/took.Seconds())
		}
		t.Logf("  total: %d frames, %.1f MB, %.1f s to decode all",
			totalFrames, float64(totalBytes)/1e6, totalDecode.Seconds())
	}
}
