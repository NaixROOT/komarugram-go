// SPDX-License-Identifier: Unlicense OR MIT

package token

import (
	"math"
	"time"
)

// @see https://m3.material.io/styles/motion/easing-and-duration/tokens-specs
const (
	DurationShort1 time.Duration = 50 * time.Millisecond
	DurationShort2 time.Duration = 100 * time.Millisecond
	DurationShort3 time.Duration = 150 * time.Millisecond
	DurationShort4 time.Duration = 200 * time.Millisecond

	DurationMedium1 time.Duration = 250 * time.Millisecond
	DurationMedium2 time.Duration = 300 * time.Millisecond
	DurationMedium3 time.Duration = 350 * time.Millisecond
	DurationMedium4 time.Duration = 400 * time.Millisecond

	DurationLong1 time.Duration = 450 * time.Millisecond
	DurationLong2 time.Duration = 500 * time.Millisecond
	DurationLong3 time.Duration = 550 * time.Millisecond
	DurationLong4 time.Duration = 600 * time.Millisecond

	DurationExtraLong1 time.Duration = 700 * time.Millisecond
	DurationExtraLong2 time.Duration = 800 * time.Millisecond
	DurationExtraLong3 time.Duration = 900 * time.Millisecond
	DurationExtraLong4 time.Duration = 1000 * time.Millisecond
)

// Easing is a cubic Bézier easing curve from (0,0) to (1,1) with the control
// points (X1,Y1) and (X2,Y2), the same as CSS cubic-bezier().
type Easing struct {
	X1, Y1, X2, Y2 float64
}

// @see https://m3.material.io/styles/motion/easing-and-duration/tokens-specs
var (
	EasingLinear               = Easing{0, 0, 1, 1}
	EasingStandard             = Easing{0.2, 0, 0, 1}
	EasingStandardAccelerate   = Easing{0.3, 0, 1, 1}
	EasingStandardDecelerate   = Easing{0, 0, 0, 1}
	EasingEmphasizedAccelerate = Easing{0.3, 0, 0.8, 0.15}
	EasingEmphasizedDecelerate = Easing{0.05, 0.7, 0.1, 1}
)

// Ease maps the animation progress t in [0, 1] to the eased progress.
func (e Easing) Ease(t float64) float64 {
	if t <= 0 {
		return 0
	}
	if t >= 1 {
		return 1
	}
	// Find the curve parameter u for which x(u) == t, then return y(u).
	// x(u) is monotonic for control points with X in [0, 1], so bisection
	// always converges; Newton's method is tried first because it is faster.
	u := t
	for range 8 {
		x := bezier(u, e.X1, e.X2) - t
		if math.Abs(x) < 1e-6 {
			return bezier(u, e.Y1, e.Y2)
		}
		d := bezierDerivative(u, e.X1, e.X2)
		if math.Abs(d) < 1e-6 {
			break
		}
		u -= x / d
	}
	lo, hi := 0.0, 1.0
	u = t
	for range 32 {
		x := bezier(u, e.X1, e.X2)
		if math.Abs(x-t) < 1e-6 {
			break
		}
		if x < t {
			lo = u
		} else {
			hi = u
		}
		u = (lo + hi) / 2
	}
	return bezier(u, e.Y1, e.Y2)
}

// bezier evaluates one coordinate of the cubic Bézier curve with the end
// points 0 and 1 and the control points p1 and p2.
func bezier(u, p1, p2 float64) float64 {
	v := 1 - u
	return 3*v*v*u*p1 + 3*v*u*u*p2 + u*u*u
}

func bezierDerivative(u, p1, p2 float64) float64 {
	v := 1 - u
	return 3*v*v*p1 + 6*v*u*(p2-p1) + 3*u*u*(1-p2)
}
