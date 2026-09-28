// SPDX-License-Identifier: Unlicense OR MIT

package tray

import "testing"

func TestIcon(t *testing.T) {
	for _, size := range []int{16, 32, 64} {
		im := Icon(size)
		if a := im.NRGBAAt(0, 0).A; a != 0 {
			t.Errorf("%dpx: corner alpha %d, want transparent", size, a)
		}
		// The middle of the bubble is white and opaque.
		if c := im.NRGBAAt(size/2, size*46/100); c.A != 255 || c.R < 250 || c.G < 250 || c.B < 250 {
			t.Errorf("%dpx: bubble %v", size, c)
		}
		// Between the bubble and the edge is the disc.
		if c := im.NRGBAAt(size/2, size*9/10); c.A != 255 || c.R > 0x70 {
			t.Errorf("%dpx: disc %v", size, c)
		}
	}
}
