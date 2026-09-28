// SPDX-License-Identifier: Unlicense OR MIT

package deviceinfo

import (
	"slices"
	"testing"
)

func TestFirmwareModel(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		product, family, board string
		want                   string
	}{
		// ASUS fills the product and the family with placeholders too long
		// to use, and the board is long as well.
		{"placeholders", "System Product Name", "To be filled by O.E.M.", "TUF GAMING B650M-E WIFI", ""},
		{"short product", "MS-7C56", "Default string", "MAG B550 TOMAHAWK (MS-7C56)", "MS-7C56"},
		{"HP words", "HP EliteBook 840 G5 Notebook PC", "", "", "HP EliteBook 840 G5"},
		{"HP limit", "HP ProBook 450 15.6 inch G9 Notebook PC Wolf Pro Security Edition", "", "", "HP ProBook 450 15.6 inch G9 Wolf"},
		{"family and board", "Standard PC (Q35 + ICH9, 2009)", "ThinkPad", "20QD", "ThinkPad 20QD"},
		{"board", "Standard PC (Q35 + ICH9, 2009)", "Some Very Long Family", "X570-A", "X570-A"},
		{"family", "Standard PC (Q35 + ICH9, 2009)", "Surface", "Some Very Long Board", "Surface"},
		// Qt counts UTF-16 code units: 14 letters and an emoji are 16, while
		// 14 Cyrillic letters are 14 whatever their bytes.
		{"UTF-16 length", "Abcdefghijklmn😀", "", "", ""},
		{"Cyrillic", "Компьютер Ивана", "", "", "Компьютер Ивана"},
		{"none", "", "", "", ""},
	} {
		if got := firmwareModel(tc.product, tc.family, tc.board); got != tc.want {
			t.Errorf("%s: firmwareModel(%q, %q, %q) = %q, want %q", tc.name, tc.product, tc.family, tc.board, got, tc.want)
		}
	}
}

func TestSimplifyModel(t *testing.T) {
	for in, want := range map[string]string{
		"MS-7C56\n":              "MS-7C56",
		"  To_Be_Filled\tBy  x ": "ToBeFilled By x",
		"a\x01b":                 "a b",
	} {
		if got := simplifyModel(in); got != want {
			t.Errorf("simplifyModel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFinalizeModel(t *testing.T) {
	for _, tc := range []struct {
		model string
		mac   bool
		want  string
	}{
		{" MS-7C56 ", false, "MS-7C56"},
		{"", false, "Desktop"},
		{"", true, "Mac"},
	} {
		if got := finalizeModel(tc.model, tc.mac); got != tc.want {
			t.Errorf("finalizeModel(%q, %v) = %q, want %q", tc.model, tc.mac, got, tc.want)
		}
	}
}

func TestWindowsSystem(t *testing.T) {
	for _, tc := range []struct {
		major, build uint32
		machine      uint16
		want         string
	}{
		{10, 26200, imageFileMachineAMD64, "Windows 11 x64"},
		{10, 22000, imageFileMachineARM64, "Windows 11 arm64"},
		{10, 19045, imageFileMachineAMD64, "Windows 10 x64"},
		{10, 19045, 0, "Windows 10"},
	} {
		if got := windowsSystem(tc.major, tc.build, tc.machine); got != tc.want {
			t.Errorf("windowsSystem(%d, %d, %#x) = %q, want %q", tc.major, tc.build, tc.machine, got, tc.want)
		}
	}
}

func TestLinuxSystem(t *testing.T) {
	for _, tc := range []struct {
		desktops                   []string
		display, libc, libcVersion string
		want                       string
	}{
		{[]string{"XFCE"}, "X11", "glibc", "2.39", "Linux XFCE X11 glibc 2.39"},
		{[]string{"ubuntu", "GNOME"}, "Wayland", "glibc", "2.39", "Linux ubuntu GNOME Wayland glibc 2.39"},
		{nil, "Xwayland", "", "1.2", "Linux Xwayland libc 1.2"},
		{nil, "", "glibc", "unknown", "Linux"},
	} {
		if got := linuxSystem(tc.desktops, tc.display, tc.libc, tc.libcVersion); got != tc.want {
			t.Errorf("linuxSystem(%q, %q, %q, %q) = %q, want %q", tc.desktops, tc.display, tc.libc, tc.libcVersion, got, tc.want)
		}
	}
}

func TestDesktopEnvironments(t *testing.T) {
	if got, want := desktopEnvironments("ubuntu: GNOME ::"), []string{"ubuntu", "GNOME"}; !slices.Equal(got, want) {
		t.Errorf("desktopEnvironments = %q, want %q", got, want)
	}
	if got := desktopEnvironments(""); len(got) != 0 {
		t.Errorf("desktopEnvironments(\"\") = %q, want none", got)
	}
}

func TestLinuxFallbackModels(t *testing.T) {
	for in, want := range map[string]string{"3": "Desktop", "10": "Laptop", "30": "Tablet", "2": "", "": ""} {
		if got := chassisModel(in); got != want {
			t.Errorf("chassisModel(%q) = %q, want %q", in, got, want)
		}
	}
	for in, want := range map[string]string{"kvm\n": "KVM", "none\n": "", "": ""} {
		if got := virtualizationModel(in); got != want {
			t.Errorf("virtualizationModel(%q) = %q, want %q", in, got, want)
		}
	}
}
