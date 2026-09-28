// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
)

type Draggable struct {
	Value float32
	drag  gesture.Drag
}

func (d *Draggable) Dragging() bool {
	return d.drag.Dragging()
}

func (d *Draggable) Layout(gtx layout.Context, axis gesture.Axis, w layout.Widget) (dimensions layout.Dimensions) {
	size := float32(gtx.Constraints.Max.X)
	if axis == gesture.Vertical {
		size = float32(gtx.Constraints.Max.Y)
	}

	for {
		e, ok := d.drag.Update(gtx.Metric, gtx.Source, axis)
		if !ok {
			break
		}
		if size > 0 {
			if e.Kind == pointer.Press || e.Kind == pointer.Drag {
				pos := e.Position.X
				if axis == gesture.Vertical {
					pos = size - e.Position.Y
				}
				d.Value = pos / size
			}
		}
	}
	if d.Value < 0 {
		d.Value = 0
	} else if d.Value > 1 {
		d.Value = 1
	}

	widgetDimensions := w(gtx)
	defer clip.Rect{Max: widgetDimensions.Size}.Op().Push(gtx.Ops).Pop()
	d.drag.Add(gtx.Ops)

	return widgetDimensions
}
