// SPDX-License-Identifier: Unlicense OR MIT

package wdk

// QuadEaseIn applies a quadratic easing function that starts slow and ends fast.
func QuadEaseIn(t float64) float64 {
	return t * t
}

// QuadEaseOut applies a quadratic easing function that starts fast and ends slow.
func QuadEaseOut(t float64) float64 {
	return -t * (t - 2)
}

// QuadEaseInOut applies a quadratic easing function that starts and ends slow with a fast middle.
func QuadEaseInOut(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -2*t*t + 4*t - 1
}

// CubicEaseIn applies a cubic easing function that starts slow and ends fast.
func CubicEaseIn(t float64) float64 {
	return t * t * t
}

// CubicEaseOut applies a cubic easing function that starts fast and ends slow.
func CubicEaseOut(t float64) float64 {
	t = t - 1
	return t*t*t + 1
}

// CubicEaseInOut applies a cubic easing function that starts and ends slow with a fast middle.
func CubicEaseInOut(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	t = t - 1
	return 4*t*t*t + 1
}
