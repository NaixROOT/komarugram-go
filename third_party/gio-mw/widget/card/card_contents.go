// SPDX-License-Identifier: Unlicense OR MIT

package card

import (
	"gio-mw/wdk/block"

	"gioui.org/layout"
	"gioui.org/widget"
)

type Child struct {
	cover  bool
	image  *widget.Image
	widget layout.Widget
}

func (c *Child) Layout(gtx layout.Context) layout.Dimensions {
	if c.cover {
		return c.layoutChild(gtx)
	}
	return block.Padding{
		Bottom: paddingBottom,
		Start:  paddingStart,
		End:    paddingEnd,
		Top:    paddingTop,
	}.Layout(gtx, c.layoutChild)
}

func (c *Child) layoutChild(gtx layout.Context) layout.Dimensions {
	if c.image != nil {
		return c.image.Layout(gtx)
	} else if c.widget != nil {
		return c.widget(gtx)
	}
	return layout.Dimensions{}
}

func Image(widget *widget.Image) *Child {
	return &Child{
		cover: true,
		image: widget,
	}
}

func Content(widget layout.Widget) *Child {
	return &Child{
		widget: widget,
	}
}

func ContentCover(widget layout.Widget) *Child {
	return &Child{
		cover:  true,
		widget: widget,
	}
}
