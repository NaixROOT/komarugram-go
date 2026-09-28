// SPDX-License-Identifier: Unlicense OR MIT

package appearance

import "testing"

func TestSchemeOfTheme(t *testing.T) {
	for name, want := range map[string]Scheme{
		"":                   Unknown,
		"Mint-Y-Purple":      Light,
		"Mint-Y-Dark-Purple": Dark,
		"Adwaita-dark":       Dark,
		"Breeze":             Light,
	} {
		if got := schemeOfTheme(name); got != want {
			t.Errorf("schemeOfTheme(%q) = %v, want %v", name, got, want)
		}
	}
}
