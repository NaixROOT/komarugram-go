// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"

	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

const (
	fieldHeight = unit.Dp(56)
	fieldRadius = unit.Dp(12)
)

// textField is a one-line outlined field with its label above it. The Material
// input widget cannot hide what is typed, which the 2FA password needs.
type textField struct {
	editor widget.Editor
}

// newTextField returns a field that shows every character as mask, if mask
// is not zero, and accepts only the characters in filter, if it is not empty.
func newTextField(mask rune, filter string) *textField {
	return &textField{editor: widget.Editor{SingleLine: true, Submit: true, Mask: mask, Filter: filter}}
}

func (f *textField) Text() string { return f.editor.Text() }

func (f *textField) Clear() { f.editor.SetText("") }

func (f *textField) Focus(gtx layout.Context) { gtx.Execute(key.FocusCmd{Tag: &f.editor}) }

// Submitted reports whether Enter was pressed in the field.
func (f *textField) Submitted(gtx layout.Context) bool {
	submitted := false
	for {
		ev, ok := f.editor.Update(gtx)
		if !ok {
			return submitted
		}
		if _, ok := ev.(widget.SubmitEvent); ok {
			submitted = true
		}
	}
}

// Layout draws the label and the field; invalid draws it in the error color.
func (f *textField) Layout(gtx layout.Context, title string, invalid bool) layout.Dimensions {
	sc := scheme(gtx)
	theme := wdk.GetMaterialTheme(gtx)
	focused := gtx.Focused(&f.editor)

	outline, width := sc.Outline, gtx.Dp(1)
	switch {
	case invalid:
		outline, width = sc.Error.Color, gtx.Dp(2)
	case focused:
		outline, width = sc.Primary.Color, gtx.Dp(2)
	}
	labelColor := sc.SurfaceVariant.OnColor
	if invalid {
		labelColor = sc.Error.Color
	} else if focused {
		labelColor = sc.Primary.Color
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Bottom: 6, Left: 4}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return label(gtx, title, token.TypestyleLabelMedium, labelColor, 1)
			})
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(fieldHeight))
			radius := gtx.Dp(fieldRadius)
			box := image.Rectangle{Max: size}
			fillRounded(gtx, sc.Surface.Color, size, radius)
			paint.FillShape(gtx.Ops, outline.AsNRGBA(), clip.Stroke{
				Path:  clip.UniformRRect(box, radius).Path(gtx.Ops),
				Width: float32(width),
			}.Op())

			pad := gtx.Dp(16)
			style := theme.Typescale[token.TypestyleBodyLarge]
			f.editor.LineHeight = style.LineHeight

			color := op.Record(gtx.Ops)
			paint.ColorOp{Color: sc.Surface.OnColor.AsNRGBA()}.Add(gtx.Ops)
			text := color.Stop()
			color = op.Record(gtx.Ops)
			paint.ColorOp{Color: sc.Primary.Color.SetOpacity(token.OpacityLevel3).AsNRGBA()}.Add(gtx.Ops)
			selection := color.Stop()

			// The line sits in the middle of the box and the editor reaches
			// down to the bottom of it, so a click anywhere below the top edge
			// of the line lands in the editor.
			top := max((size.Y-gtx.Sp(style.LineHeight))/2, 0)
			lineGtx := gtx
			lineGtx.Constraints = layout.Exact(image.Pt(max(size.X-2*pad, 0), size.Y-top))
			offset(gtx, image.Pt(pad, top), func(gtx layout.Context) layout.Dimensions {
				return f.editor.Layout(lineGtx, theme.TextShaper, style.AsRegularFont(), style.Size, text, selection)
			})
			return layout.Dimensions{Size: size}
		}),
	)
}
