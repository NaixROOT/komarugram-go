// SPDX-License-Identifier: Unlicense OR MIT

package vp9_test

import (
	"bytes"
	"image"
	"math"
	"testing"
	"time"

	"komarugram/pkg/resample"
)

// TestFrameSizedMatchesFrameAt checks that a frame scaled from its planes is
// the frame composed and then scaled, and the same frame at its own size,
// in whatever order frames are asked for.
func TestFrameSizedMatchesFrameAt(t *testing.T) {
	runtime, ctx := newRuntime(t)
	data := readSticker(t)
	sized, err := runtime.OpenSticker(ctx, "sized", data)
	if err != nil {
		t.Fatal(err)
	}
	defer sized.Close(ctx)
	// Forward, skipping frames, back to the start, and the same frame again.
	for _, f := range []int{0, 1, 5, 20, 3, 3, 29} {
		// A sticker of its own decodes the frame from the start.
		fresh, err := runtime.OpenSticker(ctx, "fresh", data)
		if err != nil {
			t.Fatal(err)
		}
		at := time.Duration(f) * fresh.File.FrameDuration
		want, err := fresh.FrameAt(ctx, at)
		if err != nil {
			t.Fatal(err)
		}
		fresh.Close(ctx)
		same, err := sized.FrameSized(ctx, at, sized.Size)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(same.Pix, want.Pix) {
			t.Fatalf("frame %d at its own size differs from FrameAt", f)
		}
		scaled := resample.Resize(want, 64, 64)
		got, err := sized.FrameSized(ctx, at, image.Pt(64, 64))
		if err != nil {
			t.Fatal(err)
		}
		var se float64
		for i := range scaled.Pix {
			d := float64(scaled.Pix[i]) - float64(got.Pix[i])
			se += d * d
		}
		if psnr := 10 * math.Log10(255*255/(se/float64(len(scaled.Pix))+1e-9)); psnr < 40 {
			t.Fatalf("frame %d scaled from its planes: PSNR %.1f dB", f, psnr)
		}
		// FrameAt after FrameSized composes the frame reached.
		again, err := sized.FrameAt(ctx, at)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(again.Pix, want.Pix) {
			t.Fatalf("frame %d composed after scaling differs", f)
		}
	}
}
