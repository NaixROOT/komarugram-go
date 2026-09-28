// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"gio-mw/token"
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/clip"
)

type CornerKind int

const (
	CornerKindDefault CornerKind = iota
	CornerKindChamfer
	CornerKindRound
)

type CornerShape struct {
	Kind        CornerKind
	Size        float32
	AdaptToSize bool
}

type CornerShapes struct {
	TopStart    CornerShape
	TopEnd      CornerShape
	BottomStart CornerShape
	BottomEnd   CornerShape
}

func (s CornerShapes) IsZero() bool {
	return s.TopStart.Kind == CornerKindDefault &&
		s.TopEnd.Kind == CornerKindDefault &&
		s.BottomStart.Kind == CornerKindDefault &&
		s.BottomEnd.Kind == CornerKindDefault
}

func FromCornerShapesToken(gtx layout.Context, shapes token.CornerShapes) CornerShapes {
	return CornerShapes{
		TopStart: CornerShape{
			Kind:        CornerKind(shapes.TopStart.Kind),
			Size:        float32(gtx.Dp(shapes.TopStart.Size)),
			AdaptToSize: shapes.TopStart.AdaptToSize,
		},
		TopEnd: CornerShape{
			Kind:        CornerKind(shapes.TopEnd.Kind),
			Size:        float32(gtx.Dp(shapes.TopEnd.Size)),
			AdaptToSize: shapes.TopEnd.AdaptToSize,
		},
		BottomStart: CornerShape{
			Kind:        CornerKind(shapes.BottomStart.Kind),
			Size:        float32(gtx.Dp(shapes.BottomStart.Size)),
			AdaptToSize: shapes.BottomStart.AdaptToSize,
		},
		BottomEnd: CornerShape{
			Kind:        CornerKind(shapes.BottomEnd.Kind),
			Size:        float32(gtx.Dp(shapes.BottomEnd.Size)),
			AdaptToSize: shapes.BottomEnd.AdaptToSize,
		},
	}
}

type ShapedRect struct {
	MinPoint f32.Point
	MaxPoint f32.Point
	Offset   float32
	Shapes   CornerShapes
}

// Resolve replaces AdaptToSize corners with explicit sizes for a box of the
// given size, whose path is inset by half the stroke width like Box does.
func (s CornerShapes) Resolve(size image.Point, strokeWidth float32) CornerShapes {
	half := (float32(min(size.X, size.Y)) - strokeWidth) / 2
	resolve := func(c CornerShape) CornerShape {
		if c.AdaptToSize {
			c.AdaptToSize = false
			c.Size = max(half, 0)
		}
		return c
	}
	return CornerShapes{
		TopStart:    resolve(s.TopStart),
		TopEnd:      resolve(s.TopEnd),
		BottomStart: resolve(s.BottomStart),
		BottomEnd:   resolve(s.BottomEnd),
	}
}

// LerpCornerShapes interpolates the corner sizes of two resolved shapes, t in
// [0, 1]. The corner kinds are taken from b once t reaches 1, from a before.
func LerpCornerShapes(a, b CornerShapes, t float32) CornerShapes {
	lerp := func(a, b CornerShape) CornerShape {
		c := a
		if t >= 1 {
			c = b
		}
		c.Size = a.Size + (b.Size-a.Size)*t
		return c
	}
	return CornerShapes{
		TopStart:    lerp(a.TopStart, b.TopStart),
		TopEnd:      lerp(a.TopEnd, b.TopEnd),
		BottomStart: lerp(a.BottomStart, b.BottomStart),
		BottomEnd:   lerp(a.BottomEnd, b.BottomEnd),
	}
}

func UniformCornerShapes(corner CornerShape) CornerShapes {
	return CornerShapes{
		BottomEnd:   corner,
		BottomStart: corner,
		TopEnd:      corner,
		TopStart:    corner,
	}
}

func (s ShapedRect) Path(gtx layout.Context) clip.PathSpec {
	rMinP := s.MinPoint.Sub(f32.Pt(s.Offset, s.Offset))
	rMaxP := s.MaxPoint.Add(f32.Pt(s.Offset, s.Offset))

	rHeight := rMaxP.Y - rMinP.Y
	rWidth := rMaxP.X - rMinP.X
	rMinDim := min(rHeight, rWidth)
	if rMinDim <= 0 {
		return clip.PathSpec{}
	}

	var corners CornerShapes
	if gtx.Locale.Direction == system.RTL {
		corners = CornerShapes{
			TopStart:    s.Shapes.TopEnd,
			TopEnd:      s.Shapes.TopStart,
			BottomStart: s.Shapes.BottomEnd,
			BottomEnd:   s.Shapes.BottomStart,
		}
	} else {
		corners = s.Shapes
	}

	ts := min(corners.TopStart.Size, rMinDim)
	te := min(corners.TopEnd.Size, rMinDim)
	be := min(corners.BottomEnd.Size, rMinDim)
	bs := min(corners.BottomStart.Size, rMinDim)
	if corners.TopStart.AdaptToSize {
		ts = rMinDim / 2
	}
	if corners.TopEnd.AdaptToSize {
		te = rMinDim / 2
	}
	if corners.BottomEnd.AdaptToSize {
		be = rMinDim / 2
	}
	if corners.BottomStart.AdaptToSize {
		bs = rMinDim / 2
	}

	var lePath clip.Path
	lePath.Begin(gtx.Ops)
	lePath.MoveTo(f32.Point{X: rMinP.X + ts, Y: rMinP.Y})
	lePath.LineTo(f32.Point{X: rMaxP.X - te, Y: rMinP.Y})
	if te > 0 {
		if corners.TopEnd.Kind == CornerKindRound {
			fPoint := f32.Point{X: rMaxP.X - te, Y: rMinP.Y + te}
			lePath.ArcTo(fPoint, fPoint, math.Pi/2)
		} else {
			lePath.LineTo(f32.Point{X: rMaxP.X, Y: rMinP.Y + te})
		}
	}
	lePath.LineTo(f32.Point{X: rMaxP.X, Y: rMaxP.Y - be})
	if be > 0 {
		if corners.BottomEnd.Kind == CornerKindRound {
			fPoint := f32.Point{X: rMaxP.X - be, Y: rMaxP.Y - be}
			lePath.ArcTo(fPoint, fPoint, math.Pi/2)
		} else {
			lePath.LineTo(f32.Point{X: rMaxP.X - be, Y: rMaxP.Y})
		}
	}
	lePath.LineTo(f32.Point{X: rMinP.X + bs, Y: rMaxP.Y})
	if bs > 0 {
		if corners.BottomStart.Kind == CornerKindRound {
			fPoint := f32.Point{X: rMinP.X + bs, Y: rMaxP.Y - bs}
			lePath.ArcTo(fPoint, fPoint, math.Pi/2)
		} else {
			lePath.LineTo(f32.Point{X: rMinP.X, Y: rMaxP.Y - bs})
		}
	}
	lePath.LineTo(f32.Point{X: rMinP.X, Y: rMinP.Y + ts})
	if ts > 0 {
		if corners.TopStart.Kind == CornerKindRound {
			fPoint := f32.Point{X: rMinP.X + ts, Y: rMinP.Y + ts}
			lePath.ArcTo(fPoint, fPoint, math.Pi/2)
		} else {
			lePath.LineTo(f32.Point{X: rMinP.X + ts, Y: rMinP.Y})
		}
	}

	lePath.Close()
	return lePath.End()
}

func (s ShapedRect) Outline(gtx layout.Context) clip.Op {
	return clip.Outline{Path: s.Path(gtx)}.Op()
}
