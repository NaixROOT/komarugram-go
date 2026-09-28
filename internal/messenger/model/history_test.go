package model

import (
	"math/rand/v2"
	"testing"
)

func TestFenwickAgainstLinear(t *testing.T) {
	vals := make([]int, 50000)
	for i := range vals {
		vals[i] = 1 + rand.IntN(800)
	}
	tree := NewHeightIndex(vals)
	for n := 0; n < 1000; n++ {
		idx := rand.IntN(len(vals))
		vals[idx] = 1 + rand.IntN(800)
		tree.Set(idx, vals[idx])
		want := int64(0)
		for _, v := range vals[:idx] {
			want += int64(v)
		}
		if tree.Prefix(idx) != want {
			t.Fatal("prefix mismatch")
		}
		i, off := tree.Find(want + int64(vals[idx]/2))
		if i != idx || off != vals[idx]/2 {
			t.Fatal(i, idx, off)
		}
	}
}
func TestEntitiesUTF16NestedAndMalformed(t *testing.T) {
	runs := TextRuns("A👋bold & link", []Entity{{Kind: "bold", Offset: 3, Length: 4}, {Kind: "italic", Offset: 4, Length: 2}, {Kind: "url", Offset: 10, Length: 4, URL: "https://telegram.org"}, {Kind: "spoiler", Offset: 2, Length: 2}, {Kind: "bold", Offset: -1, Length: 9}})
	var text string
	bold := 0
	italic := 0
	links := 0
	for _, r := range runs {
		text += r.Text
		if r.Bold {
			bold += len(r.Text)
		}
		if r.Italic {
			italic += len(r.Text)
		}
		if r.URL != "" {
			links++
		}
		if r.Spoiler {
			t.Fatal("split surrogate accepted")
		}
	}
	if text != "A👋bold & link" || bold != 4 || italic != 2 || links != 1 {
		t.Fatalf("%q bold=%d italic=%d links=%d %+v", text, bold, italic, links, runs)
	}
}
