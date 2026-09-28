// SPDX-License-Identifier: Unlicense

package chatmedia

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"

	xdraw "golang.org/x/image/draw"
)

// Tiny previews are decoded independently of the original download and blurred
// at low resolution. Work and memory stay bounded for untrusted thumbnails.
func previewImage(data []byte) image.Image {
	if len(data) == 0 || len(data) > 256<<10 {
		return nil
	}
	cfg, _, e := image.DecodeConfig(bytes.NewReader(data))
	if e != nil || cfg.Width <= 0 || cfg.Height <= 0 || cfg.Width*cfg.Height > 512*512 {
		return nil
	}
	im, _, e := image.Decode(bytes.NewReader(data))
	if e != nil {
		return nil
	}
	w := max(1, im.Bounds().Dx()*48/max(im.Bounds().Dx(), im.Bounds().Dy()))
	h := max(1, im.Bounds().Dy()*48/max(im.Bounds().Dx(), im.Bounds().Dy()))
	small := image.NewRGBA(image.Rect(0, 0, w, h))
	xdraw.BiLinear.Scale(small, small.Bounds(), im, im.Bounds(), draw.Src, nil)
	for pass := 0; pass < 2; pass++ {
		out := image.NewRGBA(small.Bounds())
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				var rr, gg, bb uint32
				count := uint32(0)
				for yy := max(0, y-2); yy < min(h, y+3); yy++ {
					for xx := max(0, x-2); xx < min(w, x+3); xx++ {
						c := small.RGBAAt(xx, yy)
						rr += uint32(c.R)
						gg += uint32(c.G)
						bb += uint32(c.B)
						count++
					}
				}
				out.SetRGBA(x, y, color.RGBA{uint8(rr / count), uint8(gg / count), uint8(bb / count), 255})
			}
		}
		small = out
	}
	return small
}
