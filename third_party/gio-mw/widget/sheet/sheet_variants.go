// SPDX-License-Identifier: Unlicense OR MIT

package sheet

import (
	"gio-mw/wdk/block"
	"gio-mw/widget/overlay"

	"gioui.org/gesture"
	"gioui.org/layout"
)

func NewSheet(position Position, widget layout.Widget) *overlay.Item {
	s := &Sheet{}
	s.content = &layout.List{Axis: layout.Vertical}
	s.handle = &gesture.Drag{}
	s.position = position
	s.overlayState = &overlayState{}
	sGravity := block.GravityBottomCenter
	if position == Side {
		sGravity = block.GravityMiddleEnd
	}
	s.overlayState.item = overlay.NewItem(func(gtx layout.Context) layout.Dimensions {
		s.update(gtx)
		return s.layout(gtx, widget)
	}, sGravity).WithScrim()
	return s.overlayState.item
}
