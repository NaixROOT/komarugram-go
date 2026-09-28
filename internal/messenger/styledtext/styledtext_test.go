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

// A line is shaped from the start of its span only; it must wrap as if the
// whole span were shaped, whatever glyphs and breaks the text has.
func TestLinePrefixWrapsAsWholeSpan(t *testing.T) {
	shaper := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	texts := []string{
		strings.Repeat("Длинный пост канала с обычным текстом, который переносится. ", 30),
		strings.Repeat("iiii llll ", 200),
		strings.Repeat("W", 900) + " tail",
		strings.Repeat("short\n", 50) + strings.Repeat("x ", 400),
		strings.Repeat("ааааааааааааааааааааааааааааааааааааааа ", 40),
		// Combining marks take no width: a line holds more runes than
		// the estimate, and is shaped again from the whole span.
		strings.Repeat("e\u0301\u0301\u0301\u0301 ", 300),
	}
	for _, content := range texts {
		for _, width := range []int{37, 120, 333, 800} {
			spans := []SpanStyle{{Size: 14, Content: "lead "}, {Size: 14, Content: content}, {Size: 14, Font: font.Font{Weight: font.Bold}, Content: " end"}}
			got, gotSize := layoutFragments(t, shaper, nil, width, spans...)
			shapeWholeSpans = true
			want, wantSize := layoutFragments(t, shaper, nil, width, spans...)
			shapeWholeSpans = false
			if gotSize != wantSize || !reflect.DeepEqual(got, want) {
				t.Fatalf("%.20q at %d: %d fragments %v, want %d %v", content, width, len(got), gotSize, len(want), wantSize)
			}
		}
	}
}
