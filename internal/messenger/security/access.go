// SPDX-License-Identifier: Unlicense OR MIT

package security

// Access says whether this process can use the TPM and, if not, what the
// user can do about it.
type Access struct {
	Kind AccessKind
	// Device is the TPM device the problem is about, where there is one.
	Device string
	// Group is the group whose members may use Device, for AccessNoGroup
	// and AccessRelogin.
	Group string
	// Detail is the error as the system reported it, for AccessFailed.
	Detail string
}

// AccessKind is why the TPM can or cannot be used.
type AccessKind int

const (
	// AccessReady means the TPM answers.
	AccessReady AccessKind = iota
	// AccessMissing means the system has no TPM 2.0 device: there is none,
	// or it is switched off in the firmware.
	AccessMissing
	// AccessNoRule means the device exists but is not given to any group of
	// users, which a distribution's TPM udev rules do (on Debian and Ubuntu,
	// the tpm-udev package, which also creates the tss group).
	AccessNoRule
	// AccessNoGroup means the user is not in Group, which may use the device.
	AccessNoGroup
	// AccessRelogin means the user has been added to Group, but this login
	// session started before that; logging in again gives it the group.
	AccessRelogin
	// AccessFailed is any other failure, described by Detail.
	AccessFailed
)

// Access returns what the last check of the TPM found.
func (m *Manager) Access() Access {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.access
}

// Recheck probes the TPM again, for when the user has changed something
// about it, and notifies subscribers of the result.
func (m *Manager) Recheck() {
	access := checkAccess(m.tpm)
	m.mu.Lock()
	m.access = access
	m.available = access.Kind == AccessReady
	m.mu.Unlock()
	m.notify()
}

// checkAccess probes tpm. A TPM with an Access(error) method explains its
// own failures, which is how tests show each of them.
func checkAccess(tpm TPM) Access {
	if tpm == nil {
		return Access{Kind: AccessMissing}
	}
	err := tpm.Probe()
	if err == nil {
		return Access{Kind: AccessReady}
	}
	if explained, ok := tpm.(interface{ Access(error) Access }); ok {
		return explained.Access(err)
	}
	if _, hardware := tpm.(hardwareTPM); !hardware {
		return Access{Kind: AccessFailed, Detail: err.Error()}
	}
	return diagnose(err)
}
