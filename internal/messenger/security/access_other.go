// SPDX-License-Identifier: Unlicense OR MIT

//go:build !linux

package security

func diagnose(err error) Access {
	return Access{Kind: AccessFailed, Detail: err.Error()}
}
