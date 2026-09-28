// SPDX-License-Identifier: Unlicense OR MIT

package scroll

import (
	"image"
	"testing"
	"time"

	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/op"
)

func TestViewportFractions(t *testing.T) {
	// 100 elements of 10px in a 200px viewport, scrolled to element 50.
	p := layout.Position{First: 50, Offset: 0, Count: 20, OffsetLast: 0, Length: 1000}
	start, end := viewportFractions(p, 100, 200)
	if start < 0.49 || start > 0.51 || end < 0.69 || end > 0.71 {
		t.Errorf("middle: got %v..%v, want about 0.5..0.7", start, end)
	}
	start, end = viewportFractions(layout.Position{Count: 5, Length: 50}, 5, 200)
	if start != 0 || end != 1 {
		t.Errorf("content shorter than viewport: got %v..%v, want 0..1", start, end)
	}
}

func frame(now time.Time, animations bool) layout.Context {
	gtx := layout.Context{Ops: new(op.Ops), Now: now, Values: map[string]any{}}
	wdk.SetAnimationsEnabled(gtx, animations)
	return gtx
}

func TestSmoothWheel(t *testing.T) {
	start := time.Unix(1000, 0)
	var l List
	l.Axis = layout.Vertical
	l.ScrollBy(frame(start, true), 100)
	l.applyWheel(frame(start.Add(50*time.Millisecond), true))
	if l.Position.Offset <= 0 || l.Position.Offset >= 100 {
		t.Fatalf("after 50ms: offset %d, want part of the way", l.Position.Offset)
	}
	// A second notch adds to what is left.
	l.ScrollBy(frame(start.Add(50*time.Millisecond), true), 100)
	l.applyWheel(frame(start.Add(time.Second), true))
	if l.Position.Offset != 200 {
		t.Errorf("after both notches: offset %d, want 200", l.Position.Offset)
	}
	// Without animations the list jumps.
	var jump List
	jump.Axis = layout.Vertical
	jump.ScrollBy(frame(start, false), 100)
	if jump.Position.Offset != 100 {
		t.Errorf("animations off: offset %d, want 100", jump.Position.Offset)
	}
}

func TestSmoothWheelLayout(t *testing.T) {
	// The first frames of an animation move by less than a pixel; that must
	// not be taken for the end of the list.
	start := time.Unix(1000, 0)
	l := List{HideScrollbar: true}
	l.Axis = layout.Vertical
	element := func(gtx layout.Context, i int) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(100, 50)}
	}
	layoutAt := func(now time.Time) {
		gtx := frame(now, true)
		gtx.Constraints = layout.Exact(image.Pt(100, 200))
		l.Layout(gtx, 100, element)
	}
	layoutAt(start)
	l.ScrollBy(frame(start, true), 100)
	for ms := 1; ms <= 300; ms += 16 {
		layoutAt(start.Add(time.Duration(ms) * time.Millisecond))
	}
	if got := l.Position.First*50 + l.Position.Offset; got != 100 {
		t.Errorf("scrolled %d px, want 100", got)
	}
}
