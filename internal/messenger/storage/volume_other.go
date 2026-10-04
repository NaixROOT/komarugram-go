// SPDX-License-Identifier: Unlicense OR MIT

//go:build !linux && !darwin && !freebsd && !windows

package storage

import "io/fs"

func volume(string) (id string, total, free int64, ok bool) { return "", 0, 0, false }

func allocated(fs.FileInfo) int64 { return 0 }
