// SPDX-License-Identifier: Unlicense OR MIT

// Package appicon gives the application's icon, assets/logo_round.png, at
// the sizes windows and the tray show it.
package appicon

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"slices"

	"komarugram/assets"
	"komarugram/pkg/resample"
)

// WindowSizes are the sizes a window's icon is given in, from the taskbar's
// 16 pixels to a task switcher's large icons.
var WindowSizes = []int{16, 24, 32, 48, 64, 128, 256}

// Images returns the icon at each size, square: the logo, which is not quite
// square, is centered on a transparent background. The logo is decoded once
// per call and not kept.
func Images(sizes ...int) []*image.NRGBA {
	logo, err := png.Decode(bytes.NewReader(assets.LogoRound))
	if err != nil {
		// The logo is built in: a test decodes it.
		panic(err)
	}
	out := make([]*image.NRGBA, len(sizes))
	// Each size is scaled from the smallest image made so far that is at
	// least twice as large, or from the logo: as sharp, and the large logo
	// is filtered once instead of for every size.
	order := make([]int, len(sizes))
	for i := range order {
		order[i] = i
	}
	slices.SortFunc(order, func(a, b int) int { return sizes[b] - sizes[a] })
	var made []*image.RGBA
	for _, i := range order {
		size := sizes[i]
		var src image.Image = logo
		for _, m := range made {
			if m.Rect.Dx() >= 2*size {
				src = m
			}
		}
		b := src.Bounds()
		fit := resample.Fit(b.Dx(), b.Dy(), image.Pt(size, size), false)
		scaled := resample.Resize(src, fit.X, fit.Y)
		made = append(made, scaled)
		im := image.NewNRGBA(image.Rect(0, 0, size, size))
		at := image.Pt((size-fit.X)/2, (size-fit.Y)/2)
		draw.Draw(im, scaled.Bounds().Add(at), scaled, image.Point{}, draw.Src)
		out[i] = im
	}
	return out
}

// Image returns the icon size pixels square.
func Image(size int) *image.NRGBA {
	return Images(size)[0]
}
