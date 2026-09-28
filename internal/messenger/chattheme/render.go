// SPDX-License-Identifier: Unlicense OR MIT
// Package chattheme decodes and paints Telegram's public theme formats.
package chattheme

import (
	"bytes"
	"compress/gzip"
	"errors"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"

	"komarugram/internal/messenger/model"
	"komarugram/pkg/resample"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

const MaxBytes = 16 << 20

func RGB(v uint32) color.NRGBA {
	return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255}
}

// Gradient uses a linear fill for two colors and smooth weighted color fields
// for three/four. phase moves the fields without changing the supplied colors.
func Gradient(colors []uint32, w, h, rotation int, phase float64) *image.RGBA {
	w = max(1, w)
	h = max(1, h)
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	if len(colors) == 0 {
		return dst
	}
	colors = colors[:min(4, len(colors))]
	angle := float64(rotation) * math.Pi / 180
	dx, dy := -math.Sin(angle), math.Cos(angle)
	span := math.Abs(dx) + math.Abs(dy)
	points := [4][2]float64{{.15, .2}, {.85, .3}, {.7, .9}, {.2, .8}}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			nx, ny := float64(x)/float64(max(1, w-1)), float64(y)/float64(max(1, h-1))
			var weights [4]float64
			switch len(colors) {
			case 1:
				weights[0] = 1
			case 2:
				t := max(0, min(1, .5+((nx-.5)*dx+(ny-.5)*dy)/span))
				weights[0], weights[1] = 1-t, t
			default:
				sum := 0.
				for i := range colors {
					px := points[i][0] + .12*math.Sin(phase+float64(i)*1.7)
					py := points[i][1] + .12*math.Cos(phase+float64(i)*1.7)
					d := (nx-px)*(nx-px) + (ny-py)*(ny-py)
					weights[i] = 1 / math.Pow(.08+d, 2)
					sum += weights[i]
				}
				for i := range colors {
					weights[i] /= sum
				}
			}
			r, g, b := 0., 0., 0.
			for i, v := range colors {
				c := RGB(v)
				r += float64(c.R) * weights[i]
				g += float64(c.G) * weights[i]
				b += float64(c.B) * weights[i]
			}
			dst.SetRGBA(x, y, color.RGBA{uint8(r + .5), uint8(g + .5), uint8(b + .5), 255})
		}
	}
	return dst
}
func Decode(data []byte) (image.Image, error) {
	if len(data) > MaxBytes {
		return nil, errors.New("wallpaper exceeds 16 MiB")
	}
	if len(data) > 2 && data[0] == 0x1f && data[1] == 0x8b {
		z, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer z.Close()
		data, err = io.ReadAll(io.LimitReader(z, MaxBytes+1))
		if err != nil {
			return nil, err
		}
		if len(data) > MaxBytes {
			return nil, errors.New("expanded wallpaper exceeds 16 MiB")
		}
	}
	if bytes.Contains(data[:min(len(data), 1024)], []byte("<svg")) {
		// oksvg handles geometry only; scripts and external image fetching are absent.
		icon, err := oksvg.ReadIconStream(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		if icon.ViewBox.W <= 0 || icon.ViewBox.H <= 0 || math.IsInf(icon.ViewBox.W, 0) || math.IsInf(icon.ViewBox.H, 0) || math.IsNaN(icon.ViewBox.W) || math.IsNaN(icon.ViewBox.H) {
			return nil, errors.New("invalid wallpaper dimensions")
		}
		scale := min(1., 1024/math.Max(icon.ViewBox.W, icon.ViewBox.H))
		w, h := max(1, int(icon.ViewBox.W*scale)), max(1, int(icon.ViewBox.H*scale))
		dst := image.NewRGBA(image.Rect(0, 0, w, h))
		icon.SetTarget(0, 0, float64(w), float64(h))
		icon.Draw(rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, dst, dst.Bounds())), 1)
		return dst, nil
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > 32<<20 {
		return nil, errors.New("wallpaper dimensions exceed limit")
	}
	im, _, err := image.Decode(bytes.NewReader(data))
	return im, err
}

// Wallpaper renders into a bounded texture off the UI goroutine.
func Wallpaper(w *model.ChatWallpaper, im image.Image) *image.RGBA {
	width, height := 512, 768
	out := Gradient(w.Colors, width, height, w.Rotation, 0)
	if len(w.Colors) == 0 {
		draw.Draw(out, out.Bounds(), image.NewUniform(RGB(0xd8e4dc)), image.Point{}, draw.Src)
	}
	if im == nil {
		return out
	}
	if w.Blur && !w.Pattern {
		im = boxBlur(scaleFit(im, 450), 12)
	}
	if w.Tile && !w.Pattern {
		for y := 0; y < height; y += im.Bounds().Dy() {
			for x := 0; x < width; x += im.Bounds().Dx() {
				draw.Draw(out, image.Rect(x, y, x+im.Bounds().Dx(), y+im.Bounds().Dy()), im, im.Bounds().Min, draw.Over)
			}
		}
		return out
	}
	b := im.Bounds()
	scale := max(float64(width)/float64(b.Dx()), float64(height)/float64(b.Dy()))
	sw, sh := max(1, int(float64(b.Dx())*scale)), max(1, int(float64(b.Dy())*scale))
	// Avoid huge intermediates for hostile aspect ratios: sample directly into the destination.
	intensity := float64(max(-100, min(100, w.Intensity))) / 100
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sx := min(b.Max.X-1, b.Min.X+int(float64(x+(sw-width)/2)/scale))
			sy := min(b.Max.Y-1, b.Min.Y+int(float64(y+(sh-height)/2)/scale))
			c := color.NRGBAModel.Convert(im.At(sx, sy)).(color.NRGBA)
			if w.Pattern {
				factor := 1 - float64(c.A)/255*intensity
				if intensity < 0 {
					factor = float64(c.A) / 255 * (-intensity)
				}
				d := out.RGBAAt(x, y)
				d.R = uint8(float64(d.R) * factor)
				d.G = uint8(float64(d.G) * factor)
				d.B = uint8(float64(d.B) * factor)
				out.SetRGBA(x, y, d)
			} else {
				out.Set(x, y, c)
			}
		}
	}
	return out
}
func scaleFit(im image.Image, n int) image.Image {
	b := im.Bounds()
	scale := min(1., float64(n)/float64(max(b.Dx(), b.Dy())))
	// Draw uses the project's streaming resampler, avoiding giant intermediates.
	return resample.Resize(im, max(1, int(float64(b.Dx())*scale)), max(1, int(float64(b.Dy())*scale)))
}
func boxBlur(src image.Image, radius int) *image.RGBA {
	b := src.Bounds()
	tmp := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	out := image.NewRGBA(tmp.Bounds())
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			var r, g, bb, a, n uint32
			for k := max(0, x-radius); k <= min(b.Dx()-1, x+radius); k++ {
				cr, cg, cb, ca := src.At(b.Min.X+k, b.Min.Y+y).RGBA()
				r += cr
				g += cg
				bb += cb
				a += ca
				n++
			}
			tmp.SetRGBA(x, y, color.RGBA{uint8(r / n >> 8), uint8(g / n >> 8), uint8(bb / n >> 8), uint8(a / n >> 8)})
		}
	}
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			var r, g, bb, a, n uint32
			for k := max(0, y-radius); k <= min(b.Dy()-1, y+radius); k++ {
				c := tmp.RGBAAt(x, k)
				r += uint32(c.R)
				g += uint32(c.G)
				bb += uint32(c.B)
				a += uint32(c.A)
				n++
			}
			out.SetRGBA(x, y, color.RGBA{uint8(r / n), uint8(g / n), uint8(bb / n), uint8(a / n)})
		}
	}
	return out
}
