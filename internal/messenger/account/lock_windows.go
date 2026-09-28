// SPDX-License-Identifier: Unlicense OR MIT

package account

import (
	"os"

	"golang.org/x/sys/windows"
)

// lockFile takes an exclusive lock on path, failing at once if another
// process holds it. The lock goes away with the process, however it ends.
func lockFile(path string) (unlock func(), err error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	handle := windows.Handle(file.Fd())
	overlapped := new(windows.Overlapped)
	flags := uint32(windows.LOCKFILE_EXCLUSIVE_LOCK | windows.LOCKFILE_FAIL_IMMEDIATELY)
	if err := windows.LockFileEx(handle, flags, 0, 1, 0, overlapped); err != nil {
		file.Close()
		return nil, err
	}
	return func() {
		_ = windows.UnlockFileEx(handle, 0, 1, 0, overlapped)
		file.Close()
	}, nil
}
