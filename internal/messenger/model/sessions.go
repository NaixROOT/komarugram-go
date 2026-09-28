// SPDX-License-Identifier: Unlicense OR MIT

package model

import (
	"context"
	"strconv"
	"strings"
	"time"
)

// Session is a device signed in to the account (Telegram's authorization).
type Session struct {
	// Current is this device; Incomplete, a login that stopped at the
	// password, which has no access to messages.
	Current, Incomplete bool
	// Official is set for Telegram's own apps.
	Official bool
	Hash     int64
	APIID    int
	// Device is the device's model, and App the application with its
	// version.
	Device, Platform, System, App string
	IP, Country                   string
	Created, Active               time.Time
}

// SessionsSource lists the account's sessions. It asks Telegram, so it is
// called off the frame.
type SessionsSource interface {
	Sessions(ctx context.Context) ([]Session, error)
}

// SessionKind is what a session's icon shows.
type SessionKind int

const (
	SessionOther SessionKind = iota
	SessionAndroid
	SessionIPhone
	SessionIPad
	SessionWindows
	SessionMac
	SessionLinux
	SessionWeb
)

// Telegram's own apps by api_id, as Telegram Desktop tells them apart
// (TypeFromEntry).
var (
	desktopAPIIDs = []int{2040, 17349, 611335}
	macAPIIDs     = []int{2834}
	androidAPIIDs = []int{5, 6, 24, 1026, 1083, 2458, 2521, 21724}
	iosAPIIDs     = []int{1, 7, 10840, 16352}
	webAPIIDs     = []int{2496, 739222, 1025907}
)

// Kind tells what kind of device s is, as Telegram Desktop does: first by
// the app, then by what the device and system say.
func (s Session) Kind() SessionKind {
	platform, device, system := strings.ToLower(s.Platform), strings.ToLower(s.Device), strings.ToLower(s.System)
	in := func(ids []int) bool {
		for _, id := range ids {
			if id == s.APIID {
				return true
			}
		}
		return false
	}
	browser := strings.Contains(device, "chrome") || strings.Contains(device, "safari") || strings.Contains(device, "firefox") || strings.Contains(device, "edg/")
	desktop := func() (SessionKind, bool) {
		switch {
		case strings.Contains(platform, "windows") || strings.Contains(system, "windows"):
			return SessionWindows, true
		case strings.Contains(platform, "macos") || strings.Contains(system, "macos"):
			return SessionMac, true
		case strings.Contains(platform, "linux") || strings.Contains(system, "linux") || strings.Contains(platform, "ubuntu") || strings.Contains(system, "ubuntu"):
			return SessionLinux, true
		}
		return SessionOther, false
	}
	switch {
	case in(androidAPIIDs):
		return SessionAndroid
	case in(desktopAPIIDs):
		if k, ok := desktop(); ok {
			return k
		}
		return SessionLinux
	case in(macAPIIDs):
		return SessionMac
	case in(webAPIIDs), browser:
		return SessionWeb
	case strings.Contains(device, "iphone"):
		return SessionIPhone
	case strings.Contains(device, "ipad"):
		return SessionIPad
	case in(iosAPIIDs):
		return SessionIPhone
	}
	if k, ok := desktop(); ok {
		return k
	}
	switch {
	case strings.Contains(platform, "android") || strings.Contains(system, "android"):
		return SessionAndroid
	case strings.Contains(platform, "ios") || strings.Contains(system, "ios"):
		return SessionIPhone
	}
	return SessionOther
}

// SessionApp is how Telegram Desktop names a session's app (ParseEntry): its
// own builds by name and a readable version, others as they call themselves.
func SessionApp(apiID int, name, version string) string {
	switch apiID {
	case 2040, 611335:
		name = "Telegram Desktop"
	case 17349:
		name = "Telegram Desktop (GitHub)"
	}
	if name == "Telegram Desktop" || name == "Telegram Desktop (GitHub)" {
		// Telegram Desktop sends its version as one number: 5012003.
		if n, err := strconv.Atoi(version); err == nil && n > 0 {
			version = strconv.Itoa(n/1000000) + "." + strconv.Itoa(n/1000%1000)
			if patch := n % 1000; patch != 0 {
				version += "." + strconv.Itoa(patch)
			}
		}
	}
	if version == "" {
		return name
	}
	return name + " " + version
}
