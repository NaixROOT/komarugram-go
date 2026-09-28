package model

import (
	"reflect"
	"testing"
)

func TestToggleReaction(t *testing.T) {
	heart, like, fire := Reaction{Emoji: "❤"}, Reaction{Emoji: "👍"}, Reaction{Emoji: "🔥"}
	with := func(r Reaction, count int, chosen bool) Reaction { r.Count, r.Chosen = count, chosen; return r }
	for _, c := range []struct {
		name   string
		before []Reaction
		toggle Reaction
		limit  int
		want   []Reaction
	}{
		{"a first reaction", nil, heart, 1, []Reaction{with(heart, 1, true)}},
		{"joining others moves it to the end", []Reaction{with(heart, 3, false), with(like, 1, false)}, heart, 1,
			[]Reaction{with(like, 1, false), with(heart, 4, true)}},
		{"taking one back", []Reaction{with(heart, 4, true), with(like, 1, false)}, heart, 1,
			[]Reaction{with(heart, 3, false), with(like, 1, false)}},
		{"taking back the last one removes it", []Reaction{with(like, 2, false), with(heart, 1, true)}, heart, 1,
			[]Reaction{with(like, 2, false)}},
		{"one allowed: the old one gives way", []Reaction{with(heart, 1, true), with(like, 5, false)}, like, 1,
			[]Reaction{with(like, 6, true)}},
		{"three allowed: the third gives way", []Reaction{with(heart, 2, true), with(like, 2, true), with(fire, 2, true)}, Reaction{Emoji: "🎉"}, 3,
			[]Reaction{with(heart, 2, true), with(like, 2, true), with(fire, 1, false), with(Reaction{Emoji: "🎉"}, 1, true)}},
		{"the paid reaction stays", []Reaction{with(Reaction{Paid: true}, 10, true), with(heart, 1, true)}, like, 1,
			[]Reaction{with(Reaction{Paid: true}, 10, true), with(like, 1, true)}},
	} {
		before := append([]Reaction(nil), c.before...)
		got := ToggleReaction(c.before, c.toggle, c.limit)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: got %+v, want %+v", c.name, got, c.want)
		}
		if !reflect.DeepEqual(before, c.before) {
			t.Errorf("%s: the list given was changed", c.name)
		}
	}
}
