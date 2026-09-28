// SPDX-License-Identifier: Unlicense OR MIT

package token

import (
	"math"
	"testing"
)

func TestEasingEndpointsAndMonotonic(t *testing.T) {
	for _, e := range []Easing{
		EasingLinear, EasingStandard, EasingStandardAccelerate,
		EasingStandardDecelerate, EasingEmphasizedAccelerate, EasingEmphasizedDecelerate,
	} {
		if got := e.Ease(0); got != 0 {
			t.Errorf("%v.Ease(0) = %v, want 0", e, got)
		}
		if got := e.Ease(1); got != 1 {
			t.Errorf("%v.Ease(1) = %v, want 1", e, got)
		}
		prev := 0.0
		for i := 1; i <= 100; i++ {
			got := e.Ease(float64(i) / 100)
			if got < prev-1e-9 {
				t.Fatalf("%v is not monotonic at %v: %v < %v", e, float64(i)/100, got, prev)
			}
			prev = got
		}
	}
}

func TestEasingValues(t *testing.T) {
	// Reference values computed with CSS cubic-bezier semantics.
	tests := []struct {
		e    Easing
		t, y float64
	}{
		{EasingLinear, 0.3, 0.3},
		{Easing{0.25, 0.1, 0.25, 1}, 0.5, 0.8024}, // CSS "ease"
		{EasingStandard, 0.5, 0.8778},
	}
	for _, tt := range tests {
		if got := tt.e.Ease(tt.t); math.Abs(got-tt.y) > 1e-3 {
			t.Errorf("%v.Ease(%v) = %v, want %v", tt.e, tt.t, got, tt.y)
		}
	}
}

func TestLerpColor(t *testing.T) {
	red := MatColor{R: 255, A: 255}
	transparent := MatColor{}
	half := LerpColor(red, transparent, 0.5)
	if half.R != 255 || half.G != 0 || half.A != 128 {
		t.Errorf("fading red to transparent: got %+v, want R=255 A=128", half)
	}
	blue := MatColor{B: 255, A: 255}
	if got := LerpColor(red, blue, 0.5); got != (MatColor{R: 128, B: 128, A: 255}) {
		t.Errorf("red to blue: got %+v", got)
	}
	if got := LerpColor(red, blue, 0); got != red {
		t.Errorf("t=0: got %+v", got)
	}
	if got := LerpColor(red, blue, 1); got != blue {
		t.Errorf("t=1: got %+v", got)
	}
}
