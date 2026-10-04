// SPDX-License-Identifier: Unlicense OR MIT

package storage

import (
	"io/fs"
	"strings"

	"golang.org/x/sys/windows"
)

func volume(path string) (id string, total, free int64, ok bool) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return "", 0, 0, false
	}
	name := make([]uint16, windows.MAX_PATH+1)
	if windows.GetVolumePathName(p, &name[0], uint32(len(name))) != nil {
		return "", 0, 0, false
	}
	var avail, all, allFree uint64
	if windows.GetDiskFreeSpaceEx(p, &avail, &all, &allFree) != nil {
		return "", 0, 0, false
	}
	return strings.ToLower(windows.UTF16ToString(name)), int64(all), int64(avail), true
}

// allocated is unknown on Windows: the length stands for it.
func allocated(fs.FileInfo) int64 { return 0 }
