// SPDX-License-Identifier: Unlicense OR MIT

package chattheme

import (
	"image"
	"image/color"
	"math"

	"golang.org/x/image/math/fixed"
	"golang.org/x/image/vector"
)

// alphaScanner draws the paths rasterx makes of an SVG into an alpha mask,
// each only within its own bounds. rasterx's ScannerGV covers the whole
// image for every path, which makes a pattern of hundreds of small paths
// take seconds.
type alphaScanner struct {
	dst    *image.Alpha
	r      vector.Rasterizer
	points []scanPoint
	min    fixed.Point26_6
	max    fixed.Point26_6
	alpha  uint16
	clip   image.Rectangle
}

type scanPoint struct {
	p     fixed.Point26_6
	start bool
}

func newAlphaScanner(dst *image.Alpha) *alphaScanner {
	s := &alphaScanner{dst: dst, alpha: 0xffff}
	s.Clear()
	return s
}

func (s *alphaScanner) add(p fixed.Point26_6, start bool) {
	s.points = append(s.points, scanPoint{p, start})
	s.min.X, s.min.Y = min(s.min.X, p.X), min(s.min.Y, p.Y)
	s.max.X, s.max.Y = max(s.max.X, p.X), max(s.max.Y, p.Y)
}

func (s *alphaScanner) Start(a fixed.Point26_6) { s.add(a, true) }
func (s *alphaScanner) Line(b fixed.Point26_6)  { s.add(b, false) }

func (s *alphaScanner) Draw() {
	if len(s.points) == 0 {
		return
	}
	bounds := image.Rect(s.min.X.Floor(), s.min.Y.Floor(), s.max.X.Ceil()+1, s.max.Y.Ceil()+1).Intersect(s.dst.Bounds())
	if s.clip != (image.Rectangle{}) {
		bounds = bounds.Intersect(s.clip)
	}
	if bounds.Empty() {
		return
	}
	s.r.Reset(bounds.Dx(), bounds.Dy())
	ox, oy := float32(bounds.Min.X), float32(bounds.Min.Y)
	// Paths are not closed: a stroke comes as the edges of its outline,
	// each started on its own, which close only all together.
	for _, q := range s.points {
		x, y := float32(q.p.X)/64-ox, float32(q.p.Y)/64-oy
		if q.start {
			s.r.MoveTo(x, y)
		} else {
			s.r.LineTo(x, y)
		}
	}
	s.r.Draw(s.dst, bounds, image.NewUniform(color.Alpha16{A: s.alpha}), image.Point{})
}

func (s *alphaScanner) GetPathExtent() fixed.Rectangle26_6 {
	return fixed.Rectangle26_6{Min: s.min, Max: s.max}
}

func (s *alphaScanner) SetBounds(w, h int) {}

func (s *alphaScanner) SetColor(c interface{}) {
	if c, ok := c.(color.Color); ok {
		_, _, _, a := c.RGBA()
		s.alpha = uint16(a)
	}
}

// SetWinding is a no-op: the vector rasterizer uses the non-zero rule.
func (s *alphaScanner) SetWinding(bool) {}

func (s *alphaScanner) Clear() {
	s.points = s.points[:0]
	s.min = fixed.Point26_6{X: math.MaxInt32, Y: math.MaxInt32}
	s.max = fixed.Point26_6{X: math.MinInt32, Y: math.MinInt32}
}

func (s *alphaScanner) SetClip(rect image.Rectangle) { s.clip = rect }
