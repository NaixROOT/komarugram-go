// SPDX-License-Identifier: Unlicense OR MIT

package deviceinfo

import (
	"strconv"
	"strings"
)

// The machine types of IsWow64Process2 that Telegram Desktop names.
const (
	imageFileMachineAMD64 = 0x8664
	imageFileMachineARM64 = 0xAA64
)

// windowsSystem names Windows by its version and the machine it runs on,
// under any emulation (SystemVersionPretty in base_info_win.cpp). Windows 11
// is Windows 10 from build 22000 on. Go runs on Windows 10 and later only,
// so the names of older versions are not needed.
func windowsSystem(major, build uint32, machine uint16) string {
	name := "Windows 10"
	if major > 10 || build >= 22000 {
		name = "Windows 11"
	}
	switch machine {
	case imageFileMachineAMD64:
		name += " x64"
	case imageFileMachineARM64:
		name += " arm64"
	}
	return name
}

// linuxSystem names Linux by the desktop environments, the display server
// and the C library (SystemVersionPretty in base_info_linux.cpp). A libc
// version that does not start with a number is left out.
//
// Telegram Desktop puts the window manager's name in place of an unknown
// desktop environment; that is not ported.
func linuxSystem(desktops []string, display, libc, libcVersion string) string {
	parts := append([]string{"Linux"}, desktops...)
	if display != "" {
		parts = append(parts, display)
	}
	if libcVersion != "" && libcVersion[0] >= '0' && libcVersion[0] <= '9' {
		if libc == "" {
			libc = "libc"
		}
		parts = append(parts, libc, libcVersion)
	}
	return strings.Join(parts, " ")
}

// desktopEnvironments splits XDG_CURRENT_DESKTOP (GetDesktopEnvironment).
func desktopEnvironments(v string) []string {
	var list []string
	for _, item := range strings.Split(v, ":") {
		if item = strings.Join(strings.Fields(item), " "); item != "" {
			list = append(list, item)
		}
	}
	return list
}

// chassisModel names a computer by its SMBIOS chassis type, the last resort
// of DeviceModelPretty on Linux (ChassisTypeToString).
func chassisModel(chassisType string) string {
	t, _ := strconv.ParseUint(chassisType, 10, 32)
	switch t {
	case 0x3, 0x4, 0x6, 0x7, 0xD:
		return "Desktop"
	case 0x8, 0x9, 0xA, 0xE:
		return "Laptop"
	case 0xB:
		return "Handset"
	case 0x11, 0x1C, 0x1D:
		return "Server"
	case 0x1E:
		return "Tablet"
	case 0x1F, 0x20:
		return "Convertible"
	}
	return ""
}

// virtualizationModel names a virtual machine by what systemd-detect-virt
// printed, upper-cased; "" for none.
func virtualizationModel(out string) string {
	v := strings.ToUpper(strings.Join(strings.Fields(out), " "))
	if v == "NONE" {
		return ""
	}
	return v
}
