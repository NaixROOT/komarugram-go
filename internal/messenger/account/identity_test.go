// SPDX-License-Identifier: Unlicense OR MIT

package account

import "testing"

func TestAppVersion(t *testing.T) {
	for _, tc := range []struct {
		goos, goarch  string
		flatpak, snap bool
		want          string
	}{
		{"windows", "amd64", false, false, "7.2.9 x64"},
		{"windows", "386", false, false, "7.2.9"},
		{"windows", "arm64", false, false, "7.2.9 arm64"},
		{"linux", "amd64", false, false, "7.2.9"},
		{"linux", "386", false, false, "7.2.9 i386"},
		{"linux", "arm64", false, false, "7.2.9 arm64"},
		{"linux", "amd64", true, true, "7.2.9 Flatpak"},
		{"linux", "amd64", false, true, "7.2.9 Snap"},
	} {
		if got := appVersion("7.2.9", tc.goos, tc.goarch, tc.flatpak, tc.snap); got != tc.want {
			t.Errorf("appVersion(%s/%s, flatpak %v, snap %v) = %q, want %q", tc.goos, tc.goarch, tc.flatpak, tc.snap, got, tc.want)
		}
	}
}
