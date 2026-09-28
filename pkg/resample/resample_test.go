package resample

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	xdraw "golang.org/x/image/draw"
)

// photo is a YCbCr image with smooth gradients, hard edges and fine
// stripes, the way a decoded JPEG arrives. Its colours stay inside the RGB
// gamut, so converting before or after filtering clamps nothing.
func photo(w, h int) *image.YCbCr { return photoRatio(w, h, image.YCbCrSubsampleRatio420) }

func photoRatio(w, h int, ratio image.YCbCrSubsampleRatio) *image.YCbCr {
	im := image.NewYCbCr(image.Rect(0, 0, w, h), ratio)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := uint8(40 + 130*x/w)
			if (x/97+y/61)%2 == 0 {
				v += 40
			}
			if y > h/2 && x%3 == 0 {
				v = 205
			}
			im.Y[im.YOffset(x, y)] = v
			ci := im.COffset(x, y)
			im.Cb[ci], im.Cr[ci] = uint8(114+28*y/h), uint8(142-28*x/w)
		}
	}
	return im
}

func toRGBA(src image.Image) *image.RGBA {
	out := image.NewRGBA(src.Bounds())
	xdraw.Draw(out, out.Bounds(), src, src.Bounds().Min, xdraw.Src)
	return out
}

func compare(t *testing.T, name string, got, want *image.RGBA, worstLimit int, meanLimit float64) {
	t.Helper()
	worst, total := 0, 0
	for i := range got.Pix {
		d := int(got.Pix[i]) - int(want.Pix[i])
		if d < 0 {
			d = -d
		}
		worst = max(worst, d)
		total += d
	}
	if mean := float64(total) / float64(len(got.Pix)); worst > worstLimit || mean > meanLimit {
		t.Errorf("%s: worst channel difference %d, mean %.3f", name, worst, mean)
	}
}

var sizes = []image.Point{{800, 533}, {320, 213}, {97, 211}, {1280, 853}, {1500, 1000}}

// The streaming filter must paint what x/image's Catmull-Rom paints; only
// float32 against float64 rounding may differ, by one level at most.
func TestMatchesXImageCatmullRom(t *testing.T) {
	src := toRGBA(photo(1280, 853))
	for _, size := range sizes {
		want := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
		xdraw.CatmullRom.Scale(want, want.Bounds(), src, src.Bounds(), xdraw.Src, nil)
		compare(t, fmt.Sprint(size), Resize(src, size.X, size.Y), want, 1, 0.25)
	}
}

// Filtering luma and chroma planes and converting afterwards is the same as
// converting first, since the filter is linear: exactly so with full chroma,
// and up to how 4:2:0 chroma is interpolated (image.YCbCr repeats each
// chroma sample over four pixels; the planes filter it).
func TestYCbCrPlanesMatchConvertingFirst(t *testing.T) {
	for _, c := range []struct {
		ratio image.YCbCrSubsampleRatio
		worst int
		mean  float64
	}{{image.YCbCrSubsampleRatio444, 1, 0.2}, {image.YCbCrSubsampleRatio420, 3, 0.2}} {
		src := photoRatio(1280, 853, c.ratio)
		rgba := toRGBA(src)
		for _, size := range sizes {
			compare(t, fmt.Sprint(c.ratio, size), Resize(src, size.X, size.Y), Resize(rgba, size.X, size.Y), c.worst, c.mean)
		}
	}
}

func TestTransparencyStaysPremultiplied(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			if (x/20+y/20)%2 == 0 {
				src.SetNRGBA(x, y, color.NRGBA{R: 255, G: 40, B: 10, A: 255})
			}
		}
	}
	got := Resize(src, 33, 33)
	for i := 0; i < len(got.Pix); i += 4 {
		a := got.Pix[i+3]
		if got.Pix[i] > a || got.Pix[i+1] > a || got.Pix[i+2] > a {
			t.Fatalf("pixel %d is not premultiplied: %v", i/4, got.Pix[i:i+4])
		}
	}
}

func TestFit(t *testing.T) {
	for _, c := range []struct {
		w, h  int
		box   image.Point
		cover bool
		want  image.Point
	}{
		{2560, 1440, image.Pt(800, 800), false, image.Pt(800, 450)},
		{2560, 1440, image.Pt(800, 800), true, image.Pt(1422, 800)},
		{640, 360, image.Pt(1920, 1080), false, image.Pt(640, 360)},
		{100, 4000, image.Pt(300, 300), false, image.Pt(8, 300)},
	} {
		if got := Fit(c.w, c.h, c.box, c.cover); got != c.want {
			t.Errorf("Fit(%d, %d, %v, %t) = %v, want %v", c.w, c.h, c.box, c.cover, got, c.want)
		}
	}
}

var sink *image.RGBA

// A 2560 px photo shown at 840 px: x/image against the streaming filter.
func BenchmarkPhotoDownscale(b *testing.B) {
	src := photo(2560, 1920)
	dw, dh := 840, 630
	b.Run("x-image", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sink = image.NewRGBA(image.Rect(0, 0, dw, dh))
			xdraw.CatmullRom.Scale(sink, sink.Bounds(), src, src.Bounds(), xdraw.Src, nil)
		}
	})
	b.Run("stream", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			sink = Resize(src, dw, dh)
		}
	})
}

// The viewer's case: a 2560 px photo shown on a 1400 px stage.
func BenchmarkViewerDownscale(b *testing.B) {
	src := photo(2560, 1707)
	b.ReportAllocs()
	for b.Loop() {
		sink = Resize(src, 1344, 896)
	}
}
