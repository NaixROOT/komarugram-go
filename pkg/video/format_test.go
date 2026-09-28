// SPDX-License-Identifier: Unlicense OR MIT

package video_test

import (
	"math"
	"path/filepath"
	"testing"
	"time"

	"komarugram/pkg/video"
)

// TestCompactMatchesRGBA checks that a clip stored as YCbCr paints the same
// picture as the same clip stored as RGBA. Chroma is subsampled, so the frames
// are close rather than identical.
func TestCompactMatchesRGBA(t *testing.T) {
	paths, err := filepath.Glob("../../videos/*.mp4")
	if err != nil || len(paths) == 0 {
		t.Skip("no ../../videos/*.mp4 to compare")
	}
	path := paths[0]

	rgba, err := video.DecodeClip(path, 240, 15, 30, video.FormatRGBA)
	if err != nil {
		t.Fatal(err)
	}
	compact, err := video.DecodeClip(path, 240, 15, 30, video.FormatYCbCr)
	if err != nil {
		t.Fatal(err)
	}
	if rgba.Len() != compact.Len() {
		t.Fatalf("frame counts differ: %d vs %d", rgba.Len(), compact.Len())
	}

	var worst float64
	for i := range rgba.Len() {
		want := rgba.FrameAt(frameTime(rgba, i))
		got := compact.FrameAt(frameTime(compact, i))
		var sum float64
		for p := range want.Pix {
			d := float64(want.Pix[p]) - float64(got.Pix[p])
			sum += d * d
		}
		rmse := math.Sqrt(sum / float64(len(want.Pix)))
		if rmse > worst {
			worst = rmse
		}
	}
	t.Logf("worst RMSE across %d frames: %.2f of 255", rgba.Len(), worst)
	if worst > 12 {
		t.Errorf("compact frames differ too much from RGBA: RMSE %.2f", worst)
	}
}

// frameTime is when frame i of the clip is on screen.
func frameTime(c *video.Clip, i int) time.Duration {
	return time.Duration(float64(i) / c.FPS * float64(time.Second))
}
