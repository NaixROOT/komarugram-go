// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"gio-mw/token"
	"math"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
)

type Editor struct {
	editor widget.Editor
}

type EditorPresentation struct {
	Color    token.MatColor
	TypeInfo token.TypeInfo
	Shaper   *text.Shaper
	// HideCaret hides the caret while the editor is empty, for example
	// until an animation around it has finished.
	HideCaret bool
}

func NewEditor(singleLine bool, enableSubmit bool) *Editor {
	return &Editor{
		editor: widget.Editor{
			SingleLine: singleLine,
			Submit:     enableSubmit,
			InputHint:  key.HintAny,
		},
	}
}

func (e *Editor) Layout(gtx layout.Context, style EditorPresentation) layout.Dimensions {
	textColorMacro := op.Record(gtx.Ops)
	color := style.Color
	if style.HideCaret && e.editor.Len() == 0 {
		// The text material also paints the caret, and an empty editor
		// paints nothing else with it.
		color = color.SetOpacity(0)
	}
	paint.ColorOp{Color: color.AsNRGBA()}.Add(gtx.Ops)
	textColor := textColorMacro.Stop()

	selectionColorMacro := op.Record(gtx.Ops)
	paint.ColorOp{Color: style.Color.SetOpacity(0.24).AsNRGBA()}.Add(gtx.Ops)
	selectionColor := selectionColorMacro.Stop()

	e.editor.LineHeight = style.TypeInfo.LineHeight
	return e.editor.Layout(
		gtx,
		style.Shaper,
		style.TypeInfo.AsRegularFont(),
		style.TypeInfo.Size,
		textColor,
		selectionColor,
	)
}

func (e *Editor) Focus(gtx layout.Context) {
	// TODO: Use key.SelectionCmd{} when fixed.
	e.editor.SetCaret(math.MaxInt, 0)
	gtx.Execute(key.FocusCmd{Tag: &e.editor})
}

func (e *Editor) Focused(gtx layout.Context) bool {
	return gtx.Focused(&e.editor)
}

func (e *Editor) Hovered() bool {
	return false
}

func (e *Editor) Submitted(gtx layout.Context) bool {
	for {
		ev, ok := e.editor.Update(gtx)
		if !ok {
			break
		}
		if ev, ok = ev.(widget.SubmitEvent); ok {
			return true
		}
	}
	return false
}

func (e *Editor) ClearText() {
	e.editor.SetText("")
}

func (e *Editor) GetText() string {
	return e.editor.Text()
}

func (e *Editor) SetText(txt string) {
	e.editor.SetText(txt)
}

func (e *Editor) SingleLine() bool {
	return e.editor.SingleLine
}
