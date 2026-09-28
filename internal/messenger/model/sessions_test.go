package model

import "testing"

func TestSessionKind(t *testing.T) {
	for _, c := range []struct {
		s    Session
		want SessionKind
	}{
		{Session{APIID: 6, Device: "Pixel 8"}, SessionAndroid},
		{Session{APIID: 2040, Platform: "Windows", System: "Windows 11"}, SessionWindows},
		{Session{APIID: 2040, Platform: "Linux"}, SessionLinux},
		{Session{APIID: 2040}, SessionLinux},
		{Session{APIID: 2834}, SessionMac},
		{Session{APIID: 10840, Device: "iPhone 15"}, SessionIPhone},
		{Session{APIID: 1, Device: "iPad Pro"}, SessionIPad},
		{Session{APIID: 2496, Device: "Chrome 140"}, SessionWeb},
		{Session{APIID: 99999, Device: "Firefox 130"}, SessionWeb},
		{Session{APIID: 99999, Platform: "Android"}, SessionAndroid},
		{Session{APIID: 99999, System: "Ubuntu 24.04"}, SessionLinux},
		{Session{APIID: 99999, Device: "Toaster"}, SessionOther},
	} {
		if got := c.s.Kind(); got != c.want {
			t.Errorf("%+v: %d, want %d", c.s, got, c.want)
		}
	}
}

func TestSessionApp(t *testing.T) {
	for _, c := range []struct {
		id            int
		name, version string
		want          string
	}{
		{2040, "Telegram Desktop", "5012003", "Telegram Desktop 5.12.3"},
		{2040, "tdesktop", "6001000", "Telegram Desktop 6.1"},
		{17349, "x", "6001000", "Telegram Desktop (GitHub) 6.1"},
		{6, "Telegram Android", "11.14.1 (5237)", "Telegram Android 11.14.1 (5237)"},
		{12345, "komarugram-go", "", "komarugram-go"},
	} {
		if got := SessionApp(c.id, c.name, c.version); got != c.want {
			t.Errorf("%d %q %q: %q, want %q", c.id, c.name, c.version, got, c.want)
		}
	}
}
