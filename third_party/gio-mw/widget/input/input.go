// SPDX-License-Identifier: Unlicense OR MIT

package input

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
)

type widgetState int

const (
	Enabled widgetState = iota
	Disabled
	Hovered
	Focused
	Error
	ErrorHovered
	ErrorFocused
)

type Input struct {
	LabelText      string
	SupportingText string
	Editor         *wdk.Editor
	LeadingIcon    wdk.IconWidget
	TrailingIcon   wdk.IconWidget
	Disabled       bool
	Error          bool
	Required       bool
	animation      animation
}

// animation holds the animated drawing parameters of an input.
type animation struct {
	initialized     bool
	float           wdk.FloatTween // 0 for the resting label, 1 for the floating one.
	label           wdk.ColorTween
	container       wdk.ColorTween
	indicator       wdk.ColorTween
	indicatorHeight wdk.FloatTween
	stateLayer      wdk.ColorTween
}

func (i *Input) getAnimation() *animation {
	a := &i.animation
	if !a.initialized {
		a.initialized = true
		a.float.Duration = token.DurationShort4
		a.stateLayer.Duration = token.DurationShort3
	}
	return a
}

func (i *Input) Update(gtx layout.Context) {
}

func (i *Input) Layout(gtx layout.Context) layout.Dimensions {
	if i.Editor == nil {
		panic("Input.Editor is required")
	}
	style := widgetStyle{
		Input:  i,
		shaper: wdk.GetTextShaper(gtx),
		theme:  BuildTheme(gtx),
	}
	return style.layout(gtx)
}

func (i *Input) getWidgetState(gtx layout.Context) widgetState {
	// States are sorted in the order of their priority.
	if i.Disabled {
		return Disabled
	} else if i.Editor.Focused(gtx) {
		if i.Error {
			return ErrorFocused
		}
		return Focused
	} else if i.Editor.Hovered() {
		if i.Error {
			return ErrorHovered
		}
		return Hovered
	}
	if i.Error {
		return Error
	}
	return Enabled
}
