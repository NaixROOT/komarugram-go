// SPDX-License-Identifier: Unlicense OR MIT

package search

import (
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/widget"
)

type widgetState int

const (
	Default widgetState = iota
	Hovered
	Focused
	Pressed
)

type Icon struct {
	Icon      wdk.IconWidget
	Label     string
	Clickable widget.Clickable
}

type Search struct {
	SupportingText string
	// LeadingIcon is a required icon to be displayed at the start of the field.
	// The optional Label causes the icon to be rendered as a button.
	LeadingIcon Icon
	// TrailingIcon is an optional icon to be displayed at the end of the field.
	// Can be used for clearing the input or to trigger an unrelated action.
	TrailingIcon Icon
	editor       *wdk.Editor
	forceState   *widgetState
}

func (s *Search) Submitted(gtx layout.Context) bool {
	if s.editor.Submitted(gtx) {
		return true
	}
	return false
}

func (s *Search) Focus(gtx layout.Context) bool {
	s.editor.Focus(gtx)
	return false
}

func (s *Search) SetText(txt string) {
	s.editor.SetText(txt)
}

func (s *Search) GetText() string {
	return s.editor.GetText()
}

func (s *Search) ClearText() {
	s.editor.ClearText()
}

// ForceState should be used only for demo purposes.
func (s *Search) ForceState(state *widgetState) *Search {
	s.forceState = state
	return s
}

func (s *Search) Layout(gtx layout.Context) layout.Dimensions {
	if s.editor == nil {
		panic("Search.editor is not defined, use Bar()")
	}
	style := widgetStyle{
		Search: s,
		shaper: wdk.GetTextShaper(gtx),
		theme:  BuildTheme(gtx),
	}
	return style.layout(gtx)
}

func (s *Search) getWidgetState(gtx layout.Context) widgetState {
	if s.forceState != nil {
		return *s.forceState
	}
	if s.editor.Focused(gtx) {
		return Focused
	} else if s.editor.Hovered() {
		return Hovered
	}
	return Default
}
