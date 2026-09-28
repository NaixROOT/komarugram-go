// SPDX-License-Identifier: Unlicense OR MIT

package security

import (
	"errors"
	"io/fs"
	"os"
	"os/user"
	"slices"
	"strconv"
	"syscall"
)

// devices are the TPM devices go-tpm opens, the resource manager first.
var devices = []string{"/dev/tpmrm0", "/dev/tpm0"}

// diagnose explains why opening the TPM failed with err. Access to the
// device is what usually stands in the way: udev gives /dev/tpmrm0 to a
// group (tss) that a user is not in by default.
func diagnose(err error) Access {
	device := ""
	var info fs.FileInfo
	for _, d := range devices {
		if i, statErr := os.Stat(d); statErr == nil {
			device, info = d, i
			break
		}
	}
	if device == "" {
		return Access{Kind: AccessMissing}
	}
	if !errors.Is(err, fs.ErrPermission) {
		return Access{Kind: AccessFailed, Device: device, Detail: err.Error()}
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Gid == 0 || info.Mode().Perm()&0o060 == 0 {
		return Access{Kind: AccessNoRule, Device: device}
	}
	gid := strconv.FormatUint(uint64(stat.Gid), 10)
	group := gid
	if g, err := user.LookupGroupId(gid); err == nil {
		group = g.Name
	}
	// The groups of the account, as /etc/group lists them now, against
	// those this process was started with.
	if u, err := user.Current(); err == nil {
		if ids, err := u.GroupIds(); err == nil && slices.Contains(ids, gid) {
			return Access{Kind: AccessRelogin, Device: device, Group: group}
		}
	}
	return Access{Kind: AccessNoGroup, Device: device, Group: group}
}
