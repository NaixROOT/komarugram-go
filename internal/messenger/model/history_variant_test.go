package model

import "testing"

func TestVariantPicksSmallestCovering(t *testing.T) {
	m := &MessageMedia{ID: "w", Width: 2560, Height: 1440, Variants: []MessageMedia{
		{ID: "m", Width: 320, Height: 180},
		{ID: "x", Width: 800, Height: 450},
		{ID: "y", Width: 1280, Height: 720},
	}}
	for _, c := range []struct {
		w, h int
		want string
	}{{100, 100, "m"}, {300, 200, "x"}, {800, 450, "x"}, {801, 300, "y"}, {2000, 100, "w"}, {0, 0, "m"}} {
		if got := m.Variant(c.w, c.h).ID; got != c.want {
			t.Errorf("Variant(%d, %d) = %s, want %s", c.w, c.h, got, c.want)
		}
	}
	old := &MessageMedia{ID: "old", Width: 1280, Height: 960}
	if old.Variant(100, 100) != old {
		t.Fatal("media without variants must stand for itself")
	}
}
