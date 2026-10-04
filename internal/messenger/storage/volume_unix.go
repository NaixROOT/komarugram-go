// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux || darwin || freebsd

package storage

import (
	"io/fs"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

func volume(path string) (id string, total, free int64, ok bool) {
	var st unix.Stat_t
	var fs unix.Statfs_t
	if unix.Stat(path, &st) != nil || unix.Statfs(path, &fs) != nil {
		return "", 0, 0, false
	}
	size := int64(fs.Bsize)
	return strconv.FormatUint(uint64(st.Dev), 10), int64(fs.Blocks) * size, int64(fs.Bavail) * size, true
}

func allocated(info fs.FileInfo) int64 {
	if st, ok := info.Sys().(*syscall.Stat_t); ok {
		return int64(st.Blocks) * 512
	}
	return 0
}
