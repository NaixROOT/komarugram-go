// SPDX-License-Identifier: Unlicense OR MIT

package security

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TestHardwareAccess reports what this machine allows, for looking at it
// rather than asserting: machines differ in TPM and group membership.
func TestHardwareAccess(t *testing.T) {
	access := checkAccess(hardwareTPM{})
	t.Logf("access %+v", access)
	if access.Kind == AccessFailed && access.Detail == "" {
		t.Error("a failure without its detail")
	}
}

func TestDiagnose(t *testing.T) {
	saved := devices
	defer func() { devices = saved }()
	dir := t.TempDir()
	device := filepath.Join(dir, "tpmrm0")
	devices = []string{device}

	if got := diagnose(fs.ErrNotExist); got.Kind != AccessMissing {
		t.Errorf("no device: %+v", got)
	}
	if err := os.WriteFile(device, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if got := diagnose(fs.ErrPermission); got.Kind != AccessNoRule || got.Device != device {
		t.Errorf("device of no group: %+v", got)
	}
	if got := diagnose(errors.New("device is not a TPM 2.0")); got.Kind != AccessFailed || got.Detail == "" {
		t.Errorf("other failure: %+v", got)
	}
	// The file belongs to the user's own group, which the account is in:
	// what a missing login looks like.
	if err := os.Chmod(device, 0o660); err != nil {
		t.Fatal(err)
	}
	if got := diagnose(fs.ErrPermission); got.Kind != AccessRelogin || got.Group == "" {
		t.Errorf("group device: %+v", got)
	}
}
