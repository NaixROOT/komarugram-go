// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"image"
)

type Box struct {
	Shape      CornerShapes
	StartPoint image.Point
	EndPoint   image.Point
	// NOTE: clip.Outline strokes are centered on the defined path, use this to inset or offset.
	StrokeWidth float32
}

func (s Box) Path(gtx layout.Context) clip.PathSpec {
	return ShapedRect{
		MinPoint: f32.Pt(float32(s.StartPoint.X), float32(s.StartPoint.Y)),
		MaxPoint: f32.Pt(float32(s.EndPoint.X), float32(s.EndPoint.Y)),
		Offset:   -s.StrokeWidth / 2,
		Shapes:   s.Shape,
	}.Path(gtx)
}

func (s Box) Outline(gtx layout.Context) clip.Op {
	return clip.Outline{Path: s.Path(gtx)}.Op()
}

func (s Box) Stroke(gtx layout.Context) clip.Op {
	return clip.Stroke{Path: s.Path(gtx), Width: s.StrokeWidth}.Op()
}
