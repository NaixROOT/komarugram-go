package styledtext

import (
	"image"
	"reflect"
	"strings"
	"testing"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
)

func layoutFragments(t *testing.T, shaper *text.Shaper, buffer *[]Cluster, width int, spans ...SpanStyle) ([]Fragment, image.Point) {
	t.Helper()
	gtx := layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: image.Pt(width, 10000)}, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}
	var fragments []Fragment
	style := Text(shaper, spans...)
	style.Clusters = buffer
	style.Decorate = func(_ layout.Context, f Fragment, draw func()) {
		// Copy: with a buffer, the next layout overwrites the clusters.
		f.Clusters = append([]Cluster(nil), f.Clusters...)
		fragments = append(fragments, f)
		draw()
	}
	return fragments, style.Layout(gtx, nil).Size
}

// Wrapped spans are laid out line by line. Every rune must belong to exactly
// one cluster, in order, and a reused cluster buffer must not change results.
func TestWrappedSpansCoverEveryRuneOnce(t *testing.T) {
	shaper := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	long := strings.Repeat("word wrap текст ", 40)
	spans := []SpanStyle{
		{Size: 14, Content: long},
		{Size: 14, Font: font.Font{Weight: font.Bold}, Content: "bold middle " + long},
		{Size: 14, Content: "line\nbreak " + long},
	}
	var all []rune
	for _, s := range spans {
		all = append(all, []rune(s.Content)...)
	}
	total := len(all)
	var buffer []Cluster
	for _, width := range []int{120, 333, 800} {
		want, wantSize := layoutFragments(t, shaper, nil, width, spans...)
		next := 0
		lines := map[int]bool{}
		for _, f := range want {
			lines[f.Bounds.Min.Y] = true
			for _, c := range f.Clusters {
				if c.Start == next+1 && all[next] == '\n' {
					// Hard line breaks are not clusters of their own.
					next++
				}
				if c.Start != next || c.End <= c.Start {
					t.Fatalf("width %d: cluster %d..%d, want start %d", width, c.Start, c.End, next)
				}
				if !c.Bounds.In(f.Bounds.Inset(-2)) {
					t.Fatalf("width %d: cluster %v outside fragment %v", width, c.Bounds, f.Bounds)
				}
				next = c.End
			}
		}
		if next != total {
			t.Fatalf("width %d: clusters end at %d of %d runes", width, next, total)
		}
		if width == 120 && len(lines) < 20 {
			t.Fatalf("expected many wrapped lines, got %d", len(lines))
		}
		for range 2 {
			got, gotSize := layoutFragments(t, shaper, &buffer, width, spans...)
			if gotSize != wantSize || !reflect.DeepEqual(got, want) {
				t.Fatalf("width %d: reused cluster buffer changed the layout", width)
			}
		}
	}
}
