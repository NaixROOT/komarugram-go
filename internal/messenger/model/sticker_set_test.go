// SPDX-License-Identifier: Unlicense OR MIT

package model

import "testing"

func TestStickerSetCreatorID(t *testing.T) {
	for _, c := range []struct {
		high, low uint64
		want      int64
	}{
		{123456789, 1, 123456789},
		{123456789, 0x003ffffe, 123456789 + 1<<31},
		{123456789, 0xff7ffffe, 123456789 + 1<<32},
		{123456789, 0xff3ffffe, 123456789 + 1<<31 + 1<<32},
		{0, 0, 0}, {123456789, 0x12000000, 0},
	} {
		id := int64(c.high<<32 | c.low)
		if got := StickerSetCreatorID(id); got != c.want {
			t.Errorf("%x: got %d want %d", id, got, c.want)
		}
	}
}
