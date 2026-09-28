// SPDX-License-Identifier: Unlicense OR MIT

package account

import "testing"

func TestParseProxy(t *testing.T) {
	for _, link := range []string{
		"tg://proxy?server=127.0.0.1&port=666&secret=dddf83f4ddaee1f1171b2e7a6e7403c60d",
		"https://t.me/proxy?server=example.org&port=443&secret=dddf83f4ddaee1f1171b2e7a6e7403c60d",
		" tg://proxy?server=127.0.0.1&port=666&secret=3d-DdruTaRaVsW6HzGY85g ",
	} {
		if _, err := ParseProxy(link); err != nil {
			t.Errorf("ParseProxy(%q): %v", link, err)
		}
	}
	for _, link := range []string{
		"",
		"tg://resolve?domain=x",
		"tg://proxy?server=127.0.0.1&port=666",
		"tg://proxy?server=127.0.0.1&port=666&secret=%%%",
		"socks5://127.0.0.1:1080",
	} {
		if _, err := ParseProxy(link); err == nil {
			t.Errorf("ParseProxy(%q) succeeded, want an error", link)
		}
	}
}
