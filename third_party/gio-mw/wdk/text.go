// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"gio-mw/token"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
)

type LabelStyle struct {
	Alignment  text.Alignment
	Color      token.MatColor
	MaxLines   int
	Typestyle  token.Typestyle
	WrapPolicy text.WrapPolicy
}

func LayoutLabel(gtx layout.Context, style LabelStyle, label string) layout.Dimensions {
	materialTheme := GetMaterialTheme(gtx)
	lTypeInfo := materialTheme.Typescale[style.Typestyle]
	lLabel := widget.Label{
		Alignment:  style.Alignment,
		MaxLines:   style.MaxLines,
		LineHeight: lTypeInfo.LineHeight,
		WrapPolicy: style.WrapPolicy,
	}

	lFont := font.Font{
		Typeface: lTypeInfo.Font,
		Style:    font.Regular,
		Weight:   lTypeInfo.Weight,
	}

	lColor := style.Color
	if style.Color.A == 0 {
		lColor = materialTheme.Scheme.Background.OnColor
	}
	lColorMacro := op.Record(gtx.Ops)
	paint.ColorOp{Color: lColor.AsNRGBA()}.Add(gtx.Ops)
	lColorCallOp := lColorMacro.Stop()

	return lLabel.Layout(gtx, materialTheme.TextShaper, lFont, lTypeInfo.Size, label, lColorCallOp)
}
