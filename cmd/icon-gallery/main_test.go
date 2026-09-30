// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"strings"
	"testing"

	"gioui.org/widget"
)

func TestEveryIconDecodes(t *testing.T) {
	seen := make(map[string]bool, len(allIcons))
	for _, ic := range allIcons {
		if seen[ic.name] {
			t.Errorf("%s is listed twice", ic.name)
		}
		seen[ic.name] = true
		if !strings.HasPrefix(ic.name, ic.category) {
			t.Errorf("%s is not in its category %s", ic.name, ic.category)
		}
		if _, err := widget.NewIcon(ic.data); err != nil {
			t.Errorf("%s: %v", ic.name, err)
		}
	}
}

func TestSearchMatchesHoweverTheNameIsWritten(t *testing.T) {
	g := &gallery{lower: make([]string, len(allIcons))}
	for i, ic := range allIcons {
		g.lower[i] = strings.ToLower(ic.name)
	}
	for _, q := range []string{"ActionSearch", "action search", "icons.ActionSearch", "ACTION_SEARCH"} {
		g.filter(normalizeQuery(q))
		found := false
		for _, i := range g.shown {
			found = found || allIcons[i].name == "ActionSearch"
		}
		if !found {
			t.Errorf("%q does not find ActionSearch; shown %d", q, len(g.shown))
		}
	}
	g.filter(normalizeQuery(""))
	if len(g.shown) != len(allIcons) {
		t.Errorf("an empty query shows %d of %d icons", len(g.shown), len(allIcons))
	}
}

func TestRowsKeepCategoriesApart(t *testing.T) {
	g := &gallery{lower: make([]string, len(allIcons))}
	for i, ic := range allIcons {
		g.lower[i] = strings.ToLower(ic.name)
	}
	g.filter("")
	g.arrange(7)
	icons := 0
	for _, r := range g.rows {
		if r.heading {
			continue
		}
		if r.n < 1 || r.n > 7 {
			t.Fatalf("a row of %d icons", r.n)
		}
		category := allIcons[g.shown[r.first]].category
		for k := range r.n {
			if c := allIcons[g.shown[r.first+k]].category; c != category {
				t.Fatalf("a row mixes %s and %s", category, c)
			}
		}
		icons += r.n
	}
	if icons != len(allIcons) {
		t.Errorf("rows hold %d of %d icons", icons, len(allIcons))
	}
}
