// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
)

func TestSpoilerUncoversOnClick(t *testing.T) {
	var s spoiler
	drawn := false
	h := &focusHarness{draw: func(gtx layout.Context) {
		drawn = false
		s.Layout(gtx, true, false, func(gtx layout.Context) layout.Dimensions {
			drawn = true
			return layout.Dimensions{Size: image.Pt(100, 20)}
		})
	}}
	h.frame()
	if s.shown {
		t.Fatal("uncovered without a click")
	}
	for _, kind := range []pointer.Kind{pointer.Press, pointer.Release} {
		h.router.Queue(pointer.Event{Kind: kind, Source: pointer.Mouse, Position: f32.Pt(10, 10), Buttons: pointer.ButtonPrimary})
		h.frame()
	}
	if !s.shown || !drawn {
		t.Fatal("a click did not uncover the value")
	}
	s.Hide()
	h.frame()
	if s.shown {
		t.Fatal("Hide left the value uncovered")
	}
}

// TestSpoilerWaveKeepsItsSpeed checks that the wave crosses a short text as
// fast as a long one, rather than taking the same time over both.
func TestSpoilerWaveKeepsItsSpeed(t *testing.T) {
	now := time.Unix(1000, 0)
	reveal := spoilerReveal{started: now, pxPerDp: 2}
	short, long := image.Pt(240, 40), image.Pt(600, 40)
	if reveal.duration(short) >= reveal.duration(long) {
		t.Fatal("a short text takes as long as a long one")
	}
	// Before either ends, the wave is as far from the click in both.
	midway := now.Add(reveal.duration(short) / 2)
	if a, b := reveal.radius(midway, short), reveal.radius(midway, long); math.Abs(float64(a-b)) > 1 {
		t.Fatalf("radius %v in the short text, %v in the long one", a, b)
	}
	if d := reveal.duration(image.Pt(4, 4)); d != spoilerMinDuration {
		t.Errorf("a tiny spoiler takes %v", d)
	}
	if d := reveal.duration(image.Pt(8000, 8000)); d != spoilerMaxDuration {
		t.Errorf("a huge spoiler takes %v", d)
	}
}

// TestSmallSpoilerIsNotSlowed checks that the lower bound of the duration
// does not slow the wave down over a single word, which is how a word came
// to open slower than a paragraph.
func TestSmallSpoilerIsNotSlowed(t *testing.T) {
	word := image.Pt(52, 23) // "секрет" at 1 px per dp.
	reveal := spoilerReveal{started: time.Unix(1000, 0), center: f32.Pt(26, 11), pxPerDp: 1}
	speed := reveal.reach(word) / reveal.duration(word).Seconds()
	if speed < spoilerSpeed*0.75 {
		t.Fatalf("the wave crosses a word at %.0f dp/s, want about %d", speed, spoilerSpeed)
	}
}
