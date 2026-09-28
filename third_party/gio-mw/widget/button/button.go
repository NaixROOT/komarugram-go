// SPDX-License-Identifier: Unlicense OR MIT

package button

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/widget"
)

type Button struct {
	bAlternativeColorScheme *AlternativeColorScheme
	bAnimation              animation
	bClickable              widget.Clickable
	bDisabled               bool
	bKind                   kind
}

func (b *Button) Disable() {
	b.bDisabled = true
}

func (b *Button) Enable() {
	b.bDisabled = false
}

func (b *Button) ClearColorScheme() {
	b.bAlternativeColorScheme = nil
}

// WithColorScheme sets a custom color scheme for text buttons.
func (b *Button) WithColorScheme(scheme *AlternativeColorScheme) *Button {
	if b.bKind != textKind {
		panic("only Text buttons accept custom color sets")
	}
	b.bAlternativeColorScheme = scheme
	return b
}

func (b *Button) Clicked(gtx layout.Context) bool {
	return b.bClickable.Clicked(gtx)
}

func (b *Button) Update(gtx layout.Context) (widget.Click, bool) {
	return b.bClickable.Update(gtx)
}

func (b *Button) Hovered(gtx layout.Context) bool {
	return b.bClickable.Hovered()
}

func (b *Button) Layout(gtx layout.Context, label string) layout.Dimensions {
	wStyle := widgetStyle{
		bAnimation: b.getAnimation(),
		bClickable: &b.bClickable,
		bDrawText:  true,
		bKind:      b.bKind,
		bLabel:     label,
		bState:     b.getWidgetState(gtx),
		bTheme:     b.getWidgetTheme(gtx),
	}
	return wStyle.layout(gtx)
}

func (b *Button) LayoutWithIcon(gtx layout.Context, label string, icon wdk.IconWidget) layout.Dimensions {
	wStyle := widgetStyle{
		bAnimation: b.getAnimation(),
		bClickable: &b.bClickable,
		bDrawIcon:  true,
		bDrawText:  true,
		bIcon:      icon,
		bKind:      b.bKind,
		bLabel:     label,
		bState:     b.getWidgetState(gtx),
		bTheme:     b.getWidgetTheme(gtx),
	}
	return wStyle.layout(gtx)
}

func (b *Button) LayoutIconOnly(gtx layout.Context, label string, icon wdk.IconWidget) layout.Dimensions {
	wStyle := widgetStyle{
		bAnimation:              b.getAnimation(),
		bClickable:              &b.bClickable,
		bDrawIcon:               true,
		bDrawText:               false,
		bIcon:                   icon,
		bKind:                   b.bKind,
		bLabel:                  label,
		bState:                  b.getWidgetState(gtx),
		bTheme:                  b.getWidgetTheme(gtx),
		bAlternativeColorScheme: b.bAlternativeColorScheme,
	}
	return wStyle.layout(gtx)
}

func (b *Button) getWidgetState(gtx layout.Context) State {
	// States are sorted in the order of their priority.
	// TODO: Add drag and drop support.
	if b.bDisabled {
		return Disabled
	} else if b.bClickable.Pressed() {
		pointer.CursorPointer.Add(gtx.Ops)
		return Pressed
	} else if b.bClickable.Hovered() {
		pointer.CursorPointer.Add(gtx.Ops)
		return Hovered
	} else if gtx.Focused(&b.bClickable) {
		return Focused
	}
	return Enabled
}

func (b *Button) getWidgetTheme(gtx layout.Context) *Theme {
	switch b.bKind {
	case elevatedKind:
		return BuildElevatedTheme(gtx)
	case filledKind:
		return BuildFilledTheme(gtx)
	case filledTonalKind:
		return BuildFilledTonalTheme(gtx)
	case outlinedKind:
		return BuildOutlinedTheme(gtx)
	case textKind:
		return BuildTextTheme(gtx)
	default:
		panic("unknown button kind")
	}
}

// animation holds the animated drawing parameters of a button.
type animation struct {
	initialized bool
	shape       wdk.FloatTween // 0 for the enabled shape, 1 for the pressed one.
	shadow      wdk.FloatTween
	container   wdk.ColorTween
	outline     wdk.ColorTween
	stateLayer  wdk.ColorTween
	label       wdk.ColorTween
	icon        wdk.ColorTween
}

func (b *Button) getAnimation() *animation {
	a := &b.bAnimation
	if !a.initialized {
		a.initialized = true
		a.shape.Duration = token.DurationShort3
		a.stateLayer.Duration = token.DurationShort3
	}
	return a
}
