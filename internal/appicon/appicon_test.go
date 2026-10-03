// SPDX-License-Identifier: Unlicense OR MIT

package appicon

import "testing"

func TestImages(t *testing.T) {
	sizes := []int{16, 32, 256}
	for i, im := range Images(sizes...) {
		size := sizes[i]
		if b := im.Bounds(); b.Dx() != size || b.Dy() != size {
			t.Fatalf("%dpx: bounds %v", size, b)
		}
		// The logo is a disc: its corners are transparent, its middle opaque.
		if a := im.NRGBAAt(0, 0).A; a != 0 {
			t.Errorf("%dpx: corner alpha %d, want transparent", size, a)
		}
		if a := im.NRGBAAt(size/2, size/2).A; a != 255 {
			t.Errorf("%dpx: middle alpha %d, want opaque", size, a)
		}
	}
}
