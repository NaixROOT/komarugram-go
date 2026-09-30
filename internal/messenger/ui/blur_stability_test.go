// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// blurredUnder draws a pattern of small marks, blurred under a clip of width
// w at a fixed place, as a menu opening does, and returns the frame.
func blurredUnder(t *testing.T, w int) image.Image {
	t.Helper()
	path := filepath.Join(t.TempDir(), "blur.png")
	renderFrames(t, image.Pt(240, 120), path, func(gtx layout.Context) {
		macro := op.Record(gtx.Ops)
		for x := 0; x < 240; x += 6 {
			for y := 0; y < 120; y += 6 {
				if (x/6+y/6)%3 == 0 {
					paint.FillShape(gtx.Ops, color.NRGBA{A: 255}, clip.Rect{Min: image.Pt(x, y), Max: image.Pt(x+3, y+4)}.Op())
				}
			}
		}
		page := macro.Stop()
		defer clip.UniformRRect(image.Rect(20, 20, 20+w, 100), 8).Push(gtx.Ops).Pop()
		layoutBackdrop(gtx, image.Pt(w, 80), image.Pt(20, 20), page)
	})
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

// The blurred backdrop of a menu that is opening keeps its place: the size
// of the layer changes with every frame, and the blurred text behind the
// menu must not swim with it.
func TestBlurDoesNotSwimWithTheSizeOfItsClip(t *testing.T) {
	for _, w := range []int{66, 68, 70, 100, 101, 102, 103} {
		before, after := blurredUnder(t, w-2), blurredUnder(t, w)
		worst := 0
		// Away from the edge that moves, where the clip cuts the pattern.
		for y := 30; y < 90; y++ {
			for x := 30; x < 20+w-30-2; x++ {
				a, _, _, _ := before.At(x, y).RGBA()
				b, _, _, _ := after.At(x, y).RGBA()
				d := int(a>>8) - int(b>>8)
				worst = max(worst, d, -d)
			}
		}
		if worst > 1 {
			t.Errorf("width %d: the blurred backdrop differs by %d/255 from that at %d", w, worst, w-2)
		}
	}
}
