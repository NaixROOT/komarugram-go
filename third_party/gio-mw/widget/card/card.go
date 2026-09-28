// SPDX-License-Identifier: Unlicense OR MIT

package card

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
)

const (
	paddingBottom       = unit.Dp(16)
	paddingStart        = unit.Dp(16)
	paddingEnd          = unit.Dp(16)
	paddingTop          = unit.Dp(16)
	SpacingBetweenCards = unit.Dp(8)
)

type Kind int

const (
	Elevated = iota
	Filled
	Outlined
)

type Card struct {
	Clickable widget.Clickable
	Kind      Kind
	Data      any
}

func (c *Card) Layout(gtx layout.Context, children ...*Child) layout.Dimensions {
	cStyle := widgetStyle{
		cCard: c,
	}
	switch c.Kind {
	case Elevated:
		cStyle.cTheme = BuildElevatedTheme(gtx)
	case Outlined:
		cStyle.cTheme = BuildOutlinedTheme(gtx)
	case Filled:
		cStyle.cTheme = BuildFilledTheme(gtx)
	default:
		panic("unknown card Kind")
	}
	return cStyle.mainLayout(gtx, children...)
}
