// SPDX-License-Identifier: Unlicense OR MIT

package deviceinfo

import (
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// model reads what the firmware says of the computer from the registry.
func model() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\BIOS`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()
	value := func(name string) string {
		s, _, _ := k.GetStringValue(name)
		return simplifyModel(s)
	}
	return firmwareModel(value("SystemProductName"), value("SystemFamily"), value("BaseBoardProduct"))
}

func system() string {
	v := windows.RtlGetVersion()
	return windowsSystem(v.MajorVersion, v.BuildNumber, nativeMachine())
}

// nativeMachine is the machine type of the computer, whatever this process
// is built for; 0 when unknown.
func nativeMachine() uint16 {
	var process, native uint16
	if windows.IsWow64Process2(windows.CurrentProcess(), &process, &native) == nil {
		return native
	}
	// Before Windows 10 1511 only a 32-bit process on x64 can tell.
	var wow64 bool
	if windows.IsWow64Process(windows.CurrentProcess(), &wow64) == nil && wow64 {
		return imageFileMachineAMD64
	}
	return 0
}
