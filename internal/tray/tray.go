// SPDX-License-Identifier: Unlicense OR MIT

// Package tray shows the application's icon in the system tray: a
// StatusNotifierItem on Linux and FreeBSD desktops, a notification area icon
// on Windows. The icon opens the application on a click and offers a menu.
package tray

import (
	"errors"
	"image"
	"image/color"
	"math"
)

// ErrUnsupported is returned by Start where no tray is implemented.
var ErrUnsupported = errors.New("tray: not supported on this platform")

// Item is an entry of the icon's menu.
type Item struct {
	Label string
	// Action runs on a goroutine of its own when the item is chosen, with
	// the activation token the desktop provided, if any.
	Action func(token string)
	// Separator draws a line instead of an item; Label and Action are unused.
	Separator bool
}

type Options struct {
	// ID names the application to the desktop; it should not change.
	ID    string
	Title string
	// Activate runs on a goroutine of its own when the icon is clicked. On
	// Wayland, token is the XDG activation token that lets the application
	// raise a window; it is empty where there is none.
	Activate func(token string)
	Items    []Item
}

// Icon returns the application icon drawn size pixels square: a chat bubble
// on a colored disc.
func Icon(size int) *image.NRGBA {
	const samples = 4
	disc := color.NRGBA{R: 0x67, G: 0x50, B: 0xa4, A: 0xff}
	im := image.NewNRGBA(image.Rect(0, 0, size, size))
	s := float64(size)
	for y := range size {
		for x := range size {
			var discCover, bubbleCover int
			for sy := range samples {
				for sx := range samples {
					// The sample point in units of the icon, 0 to 1.
					px := (float64(x) + (float64(sx)+0.5)/samples) / s
					py := (float64(y) + (float64(sy)+0.5)/samples) / s
					if math.Hypot(px-0.5, py-0.5) > 0.5 {
						continue
					}
					discCover++
					if inBubble(px, py) {
						bubbleCover++
					}
				}
			}
			if discCover == 0 {
				continue
			}
			// White bubble over the disc, the edge of the disc antialiased by
			// alpha and that of the bubble by blending.
			t := float64(bubbleCover) / float64(discCover)
			mix := func(c uint8) uint8 { return uint8(math.Round(float64(c) + (255-float64(c))*t)) }
			im.SetNRGBA(x, y, color.NRGBA{
				R: mix(disc.R), G: mix(disc.G), B: mix(disc.B),
				A: uint8(math.Round(255 * float64(discCover) / samples / samples)),
			})
		}
	}
	return im
}

// inBubble reports whether a point of the unit square is inside the chat
// bubble: an ellipse with a tail at its lower left.
func inBubble(x, y float64) bool {
	const cx, cy, rx, ry = 0.5, 0.46, 0.27, 0.21
	dx, dy := (x-cx)/rx, (y-cy)/ry
	if dx*dx+dy*dy <= 1 {
		return true
	}
	// The tail is the triangle (0.3, 0.55), (0.44, 0.62), (0.26, 0.76).
	return inTriangle(x, y, 0.3, 0.55, 0.44, 0.62, 0.26, 0.76)
}

func inTriangle(x, y, ax, ay, bx, by, cx, cy float64) bool {
	side := func(x1, y1, x2, y2 float64) float64 { return (x-x2)*(y1-y2) - (x1-x2)*(y-y2) }
	d1, d2, d3 := side(ax, ay, bx, by), side(bx, by, cx, cy), side(cx, cy, ax, ay)
	negative := d1 < 0 || d2 < 0 || d3 < 0
	positive := d1 > 0 || d2 > 0 || d3 > 0
	return !(negative && positive)
}
