// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"gioui.org/f32"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"math"
)

type Arc struct {
	Center     f32.Point
	Diameter   float32
	StartAngle float32
	SweepAngle float32
}

func (a Arc) Path(gtx layout.Context) clip.PathSpec {
	aRadius := a.Diameter / 2.0
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(a.Center)

	sweepAngleValue := a.SweepAngle
	if gtx.Locale.Direction == system.RTL {
		sweepAngleValue = -a.SweepAngle
	}

	startAngle := math.Pi * 2 * a.StartAngle / 360
	startY, startX := math.Sincos(float64(startAngle))
	if gtx.Locale.Direction == system.RTL {
		startX = -startX
	}
	sPoint := a.Center.Add(f32.Point{X: float32(startX) * aRadius, Y: float32(startY) * -aRadius})
	p.MoveTo(sPoint)

	endAngle := math.Pi * 2 * sweepAngleValue / 360
	p.ArcTo(a.Center, a.Center, endAngle)

	return p.End()
}
