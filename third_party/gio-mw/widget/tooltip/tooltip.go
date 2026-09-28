// SPDX-License-Identifier: Unlicense OR MIT

package tooltip

import (
	"gioui.org/layout"
	"gioui.org/op"
)

type Tooltip struct {
	SupportingText string
}

func (t *Tooltip) Layout(gtx layout.Context, targetWidget layout.Widget, displayTooltip bool) layout.Dimensions {
	targetDimensions := targetWidget(gtx)

	// TODO: Refactor to use overlay.
	// TODO: Refactor to use a hoverable and focusable target struct.
	if displayTooltip {
		macroOp := op.Record(gtx.Ops)
		style := widgetStyle{
			Tooltip: t,
			theme:   BuildPlainTheme(gtx),
		}
		style.layout(gtx, targetDimensions)
		callOp := macroOp.Stop()
		op.Defer(gtx.Ops, callOp)
	}

	return targetDimensions
}
