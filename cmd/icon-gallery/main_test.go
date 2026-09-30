// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"image"
	"strings"
	"testing"

	"gio-mw/exp/appearance"
	"gio-mw/wdk"
	"gio-mw/widget/button"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
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

func TestThemeModeOverridesTheSystem(t *testing.T) {
	for _, c := range []struct {
		mode   themeMode
		system appearance.Scheme
		dark   bool
	}{
		{themeSystem, appearance.Light, false},
		{themeSystem, appearance.Dark, true},
		{themeSystem, appearance.Unknown, false},
		{themeLight, appearance.Dark, false},
		{themeDark, appearance.Light, true},
		{themeDark, appearance.Unknown, true},
	} {
		if got := c.mode.isDark(c.system); got != c.dark {
			t.Errorf("%s with the system's scheme %d: dark is %v, want %v", themeLabels[c.mode], c.system, got, c.dark)
		}
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

// The button is measured once, in the theme of its first frame: whichever
// that is, the room must be that of the widest label in either theme.
func TestThemeButtonKeepsItsWidth(t *testing.T) {
	for _, darkFirst := range []bool{false, true} {
		checkThemeButtonWidth(t, []bool{darkFirst, !darkFirst})
	}
}

func checkThemeButtonWidth(t *testing.T, themes []bool) {
	g := &gallery{themeButton: button.Text(), themeMeasure: button.Text()}
	width := -1
	for _, dark := range themes {
		for mode := range themeModes {
			gtx := layout.Context{
				Ops:         new(op.Ops),
				Metric:      unit.Metric{PxPerDp: 1.5, PxPerSp: 1.5},
				Constraints: layout.Constraints{Max: image.Pt(1000, 200)},
				Values:      map[string]any{},
			}
			wdk.InitMaterialThemeInContext(gtx, g.theme(gtx, dark))
			g.mode = mode
			w := g.layoutThemeButton(gtx).Size.X
			if width == -1 {
				width = w
			}
			if w != width {
				t.Errorf("%s, dark %v: %d px wide, not %d", themeLabels[mode], dark, w, width)
			}
		}
	}
}
