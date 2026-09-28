// SPDX-License-Identifier: Unlicense OR MIT

//go:build unix

package account

import (
	"os"

	"golang.org/x/sys/unix"
)

// lockFile takes an exclusive lock on path, failing at once if another
// process holds it. The lock goes away with the process, however it ends.
func lockFile(path string) (unlock func(), err error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		file.Close()
		return nil, err
	}
	return func() {
		_ = unix.Flock(int(file.Fd()), unix.LOCK_UN)
		file.Close()
	}, nil
}
