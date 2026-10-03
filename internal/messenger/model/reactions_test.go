package model

import (
	"reflect"
	"slices"
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

func TestShownReactions(t *testing.T) {
	heart, like, fire, party := Reaction{Emoji: "❤"}, Reaction{Emoji: "👍"}, Reaction{Emoji: "🔥"}, Reaction{Emoji: "🎉"}
	custom := Reaction{DocumentID: 7}
	with := func(r Reaction, count int, chosen bool) Reaction { r.Count, r.Chosen = count, chosen; return r }
	order := []Reaction{like, heart, fire, party}
	rank := func(r Reaction) (int, bool) {
		for i, one := range order {
			if one.Same(r) {
				return i, true
			}
		}
		return 0, false
	}
	list := []Reaction{with(custom, 2, false), with(heart, 2, false), with(fire, 5, true), with(Reaction{Paid: true}, 1, false), with(like, 2, true), with(party, 4, false)}
	want := []Reaction{with(Reaction{Paid: true}, 1, false), with(fire, 5, true), with(party, 4, false), with(heart, 2, false), with(custom, 2, false), with(like, 2, true)}
	if got := ShownReactions(nil, list, rank); !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	// Without ranks, by emoji.
	if got := ShownReactions(nil, []Reaction{with(like, 1, false), with(heart, 1, false)}, nil); got[0].Emoji != "❤" {
		t.Errorf("unranked: got %+v", got)
	}

	// The account's own vote never moves a reaction: the bug was a reaction
	// that went to the end when chosen, and back when Telegram answered.
	shown := func(list []Reaction) []Reaction {
		out := ShownReactions(nil, list, rank)
		for i := range out {
			out[i].Count, out[i].Chosen = 0, false
		}
		return out
	}
	for _, limit := range []int{1, 3} {
		before := []Reaction{with(heart, 3, false), with(like, 3, false), with(fire, 1, false), with(party, 2, true)}
		for _, r := range []Reaction{heart, like, fire, party} {
			after := ToggleReaction(before, r, limit)
			if !reflect.DeepEqual(shown(after), shown(before)) {
				t.Errorf("limit %d, toggling %s: %+v became %+v", limit, r.Emoji, shown(before), shown(after))
			}
			// And as Telegram sends it back, by votes.
			server := append([]Reaction(nil), after...)
			slices.SortStableFunc(server, func(a, b Reaction) int { return b.Count - a.Count })
			if !reflect.DeepEqual(shown(server), shown(before)) {
				t.Errorf("limit %d, toggling %s: Telegram's %+v shown as %+v", limit, r.Emoji, server, shown(server))
			}
			before = after
		}
	}
}
