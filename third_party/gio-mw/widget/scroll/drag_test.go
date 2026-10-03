// SPDX-License-Identifier: Unlicense OR MIT

package scroll

import (
	"image"
	"math"
	"testing"
	"time"

	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

var cachedTheme *token.Theme

func testTheme(gtx layout.Context) *token.Theme {
	if cachedTheme == nil {
		cachedTheme = defaults.NewTheme(gtx, schemes.SchemeBaselineLight())
	}
	return cachedTheme
}

// harness runs a List of n rows of rowHeight in a 300x760 window, or a
// horizontal list of n columns in a 760x60 window.
type harness struct {
	r          input.Router
	l          List
	now        time.Time
	n          int
	row        int
	pressed    bool
	viewport   int
	horizontal bool
}

func newHarness(n, rowHeight int) *harness {
	h := &harness{now: time.Unix(1000, 0), n: n, row: rowHeight, viewport: 760}
	h.l.Axis = layout.Vertical
	h.frame()
	return h
}

func newHorizontalHarness(n, columnWidth int) *harness {
	h := &harness{now: time.Unix(1000, 0), n: n, row: columnWidth, viewport: 760, horizontal: true}
	h.l.Axis = layout.Horizontal
	h.frame()
	return h
}

func (h *harness) frame() {
	size := image.Pt(300, h.viewport)
	element := image.Pt(300, h.row)
	if h.horizontal {
		size = image.Pt(h.viewport, 60)
		element = image.Pt(h.row, 60)
	}
	gtx := layout.Context{
		Ops: new(op.Ops), Source: h.r.Source(), Now: h.now,
		Constraints: layout.Exact(size),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Values:      map[string]any{},
	}
	wdk.InitMaterialThemeInContext(gtx, testTheme(gtx))
	h.l.Layout(gtx, h.n, func(gtx layout.Context, i int) layout.Dimensions {
		return layout.Dimensions{Size: element}
	})
	h.r.Frame(gtx.Ops)
	h.now = h.now.Add(16 * time.Millisecond)
}

func (h *harness) event(kind pointer.Kind, x, y float32) {
	switch kind {
	case pointer.Press:
		h.pressed = true
	case pointer.Release:
		h.pressed = false
	}
	e := pointer.Event{Kind: kind, Source: pointer.Mouse, Position: f32.Pt(x, y)}
	if h.pressed || kind == pointer.Release {
		e.Buttons = pointer.ButtonPrimary
	}
	h.r.Queue(e)
	h.frame()
}

func (h *harness) scrolled() int {
	return h.l.Position.First*h.row + h.l.Position.Offset
}

// TestThumbFollowsPointer drags the thumb in small steps: the thumb must move
// exactly with the pointer and the list by the matching share of its length,
// for short and long lists alike.
func TestThumbFollowsPointer(t *testing.T) {
	for _, tc := range []struct {
		name string
		n    int
	}{{"short", 15}, {"long", 500}} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(tc.n, 64)
			h.event(pointer.Move, 150, 300)
			x := float32(295)
			grabY := float32(h.l.thumbStart + 10)
			h.event(pointer.Move, x, grabY)
			h.event(pointer.Press, x, grabY)
			startThumb := h.l.thumbStart
			scrollable := float64(tc.n*64 - h.viewport)
			for step := 1; step <= 60; step++ {
				h.event(pointer.Move, x, grabY+float32(step))
				wantThumb := min(startThumb+step, h.l.inset+h.l.free)
				if h.l.thumbStart != wantThumb {
					t.Fatalf("step %d: thumb at %d, want %d", step, h.l.thumbStart, wantThumb)
				}
				wantScroll := float64(wantThumb-h.l.inset) / float64(h.l.free) * scrollable
				if got := float64(h.scrolled()); math.Abs(got-wantScroll) > 1 {
					t.Fatalf("step %d: scrolled %v px, want %.1f", step, got, wantScroll)
				}
			}
			h.event(pointer.Release, x, grabY+60)
			// The released thumb stays where it was dragged.
			if h.l.thumbStart < startThumb+59 || h.l.thumbStart > startThumb+61 {
				t.Errorf("after release: thumb at %d, want about %d", h.l.thumbStart, startThumb+60)
			}
			// Leaving the list hides the scrollbar after a while.
			h.event(pointer.Move, 400, 300)
			for range 150 {
				h.frame()
			}
			if h.l.hovered || h.l.trackHovered || h.l.dragging {
				t.Errorf("pointer still over the list: hovered %v track %v dragging %v", h.l.hovered, h.l.trackHovered, h.l.dragging)
			}
		})
	}
}

// TestTrackClick checks that a click on the track moves the list by a page
// at once while the thumb glides to its new place.
func TestTrackClick(t *testing.T) {
	h := newHarness(500, 64)
	h.event(pointer.Move, 150, 300)
	thumbBefore := h.l.thumbStart
	x, y := float32(295), float32(h.l.thumbStart+h.l.thumbLength+20)
	h.event(pointer.Move, x, y)
	h.event(pointer.Press, x, y)
	wantScroll := int(float32(h.viewport) * pageFraction)
	if got := h.scrolled(); got < wantScroll-1 || got > wantScroll+1 {
		t.Fatalf("after the click: scrolled %d, want a page of %d at once", got, wantScroll)
	}
	h.event(pointer.Release, x, y)
	mid := h.l.thumbStart
	for range 20 {
		h.frame()
	}
	final := h.l.thumbStart
	if !(mid > thumbBefore && mid < final) {
		t.Errorf("thumb did not glide: before %d, soon after %d, at rest %d", thumbBefore, mid, final)
	}
}

// TestHugeListThumb checks that the thumb of a very long list keeps its
// minimum length, so it can still be hit, and still follows the pointer.
func TestHugeListThumb(t *testing.T) {
	h := newHarness(100000, 64)
	h.event(pointer.Move, 150, 300)
	if h.l.thumbLength != 32 {
		t.Fatalf("thumb length %d, want the minimum of 32", h.l.thumbLength)
	}
	x, y := float32(295), float32(h.l.thumbStart+16)
	h.event(pointer.Move, x, y)
	h.event(pointer.Press, x, y)
	h.event(pointer.Move, x, y+100)
	h.event(pointer.Release, x, y+100)
	if h.l.thumbStart != h.l.inset+100 {
		t.Errorf("thumb at %d, want %d", h.l.thumbStart, h.l.inset+100)
	}
	want := 100 / float64(h.l.free) * float64(100000*64-h.viewport)
	if got := float64(h.scrolled()); math.Abs(got-want) > 64 {
		t.Errorf("scrolled %.0f px, want about %.0f", got, want)
	}
}

// scroll sends a touchpad's scroll and lets it settle.
func (h *harness) scroll(x, y float32) {
	h.send(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(300, 30), Scroll: f32.Pt(x, y)})
	for range 20 {
		h.frame()
	}
}

// wheel sends a wheel's notches and lets them settle.
func (h *harness) wheel(x, y float32) {
	h.send(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(300, 30), Scroll: f32.Pt(x, y), Wheel: true})
	for range 20 {
		h.frame()
	}
}

func (h *harness) send(e pointer.Event) {
	h.r.Queue(e)
}

// fixedScales sets both scales to 1 for the test, whatever the platform.
func fixedScales(t *testing.T) {
	wheel, touchpad := WheelScale, TouchpadScale
	WheelScale, TouchpadScale = 1, 1
	t.Cleanup(func() { WheelScale, TouchpadScale = wheel, touchpad })
}

// A wheel's notch glides; a touchpad's scroll, however large, moves the
// list at once, scaled by TouchpadScale, as a fast swipe or the kinetic
// scrolling after it sends large distances too.
func TestWheelGlidesTouchpadFollows(t *testing.T) {
	fixedScales(t)
	h := newHarness(100, 64)
	h.event(pointer.Move, 150, 300)
	h.send(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(150, 300), Scroll: f32.Pt(0, 100), Wheel: true})
	for range 4 {
		h.frame()
	}
	if got := h.scrolled(); got <= 0 || got >= 100 {
		t.Fatalf("a notch scrolled %d px in 50 ms, want part of the way", got)
	}
	for range 20 {
		h.frame()
	}
	if got := h.scrolled(); got != 100 {
		t.Fatalf("a notch scrolled %d px, want 100", got)
	}
	TouchpadScale = 2.5
	h.send(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(150, 300), Scroll: f32.Pt(0, 80)})
	h.frame()
	if got := h.scrolled(); got != 300 {
		t.Fatalf("a touchpad's 80 px scrolled to %d at once, want 300", got)
	}
	// The wheel is not scaled as the touchpad is.
	h.send(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(150, 300), Scroll: f32.Pt(0, 100), Wheel: true})
	for range 20 {
		h.frame()
	}
	if got := h.scrolled(); got != 400 {
		t.Fatalf("a notch after the touchpad scrolled to %d, want 400", got)
	}
}

// TestHorizontal checks that a horizontal list scrolls with the vertical
// wheel and horizontal touchpad scrolling, and that its thumb, along the
// bottom edge, follows the pointer.
func TestHorizontal(t *testing.T) {
	fixedScales(t)

	h := newHorizontalHarness(100, 64)
	h.event(pointer.Move, 300, 30)
	h.wheel(0, 100) // A wheel notch.
	if got := h.scrolled(); got != 100 {
		t.Fatalf("vertical wheel: scrolled %d, want 100", got)
	}
	h.scroll(-20, 0) // A touchpad swipe to the left.
	if got := h.scrolled(); got != 80 {
		t.Fatalf("horizontal scroll: scrolled %d, want 80", got)
	}

	// Drag the thumb, which sits along the bottom edge.
	x, y := float32(h.l.thumbStart+10), float32(55)
	h.event(pointer.Move, x, y)
	h.event(pointer.Press, x, y)
	start := h.l.thumbStart
	for step := 1; step <= 30; step++ {
		h.event(pointer.Move, x+float32(step), y)
		if h.l.thumbStart != start+step {
			t.Fatalf("step %d: thumb at %d, want %d", step, h.l.thumbStart, start+step)
		}
	}
	h.event(pointer.Release, x+30, y)
}
