// SPDX-License-Identifier: Unlicense OR MIT

package miniapp

import "testing"

// The fragment carries the fields this package knows in a fixed order, then
// whatever else the client was given.
func TestFragmentKeepsExtra(t *testing.T) {
	p := Params{InitData: "a%3Db", Version: "7.0", Platform: "tdesktop", ThemeParams: "%7B%7D", Extra: "tgWebAppStartParam=go&tgWebAppBotInline=1"}
	want := "tgWebAppData=a%3Db&tgWebAppVersion=7.0&tgWebAppPlatform=tdesktop&tgWebAppThemeParams=%7B%7D&tgWebAppStartParam=go&tgWebAppBotInline=1"
	if got := p.fragment(); got != want {
		t.Fatalf("fragment %q, want %q", got, want)
	}
	if got := (Params{Version: "7.0"}).fragment(); got != "tgWebAppVersion=7.0" {
		t.Fatalf("fragment %q without extra", got)
	}
}
