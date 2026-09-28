// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android

package powersave

import (
	"github.com/godbus/dbus/v5"
)

const (
	upowerDest        = "org.freedesktop.UPower"
	upowerPath        = "/org/freedesktop/UPower"
	upowerDisplayPath = "/org/freedesktop/UPower/devices/DisplayDevice"
	upowerDeviceIface = "org.freedesktop.UPower.Device"
	upowerTypeBattery = 2

	portalDest     = "org.freedesktop.portal.Desktop"
	portalPath     = "/org/freedesktop/portal/desktop"
	portalSettings = "org.freedesktop.portal.Settings"

	propertiesIface   = "org.freedesktop.DBus.Properties"
	propertiesChanged = "PropertiesChanged"
)

// powerProfiles lists the power-profiles-daemon D-Bus names, newest first.
var powerProfiles = []struct{ dest, path string }{
	{"org.freedesktop.UPower.PowerProfiles", "/org/freedesktop/UPower/PowerProfiles"},
	{"net.hadess.PowerProfiles", "/net/hadess/PowerProfiles"},
}

// Portal settings that turn animations off.
var (
	gnomeAnimations = [2]string{"org.gnome.desktop.interface", "enable-animations"}
	kdeAnimations   = [2]string{"org.kde.kdeglobals.KDE", "AnimationDurationFactor"}
)

func startPlatform(m *Monitor) func() {
	var conns []*dbus.Conn
	if conn, err := dbus.ConnectSystemBus(); err == nil {
		conns = append(conns, conn)
		watchSystem(m, conn)
	}
	if conn, err := dbus.ConnectSessionBus(); err == nil {
		conns = append(conns, conn)
		watchPortal(m, conn)
	}
	return func() {
		for _, conn := range conns {
			conn.Close()
		}
	}
}

// watchSystem reads the power profile and battery and follows their changes.
func watchSystem(m *Monitor, conn *dbus.Conn) {
	read := func() {
		profile, profileOK := readPowerProfile(conn)
		onBattery, batteryOK := readOnBattery(conn)
		percent := readBatteryPercent(conn)
		m.update(func(s *State) {
			if profileOK {
				s.PowerSaver = profile == "power-saver"
			}
			if batteryOK {
				s.OnBattery = onBattery
			}
			s.BatteryPercent = percent
		})
	}
	read()

	paths := []dbus.ObjectPath{upowerPath, upowerDisplayPath}
	for _, p := range powerProfiles {
		paths = append(paths, dbus.ObjectPath(p.path))
	}
	for _, path := range paths {
		_ = conn.AddMatchSignal(
			dbus.WithMatchObjectPath(path),
			dbus.WithMatchInterface(propertiesIface),
			dbus.WithMatchMember(propertiesChanged),
		)
	}
	signals := make(chan *dbus.Signal, 16)
	conn.Signal(signals)
	go func() {
		// The channel is closed when the connection is closed.
		for range signals {
			read()
		}
	}()
}

func readPowerProfile(conn *dbus.Conn) (string, bool) {
	for _, p := range powerProfiles {
		v, err := conn.Object(p.dest, dbus.ObjectPath(p.path)).GetProperty(p.dest + ".ActiveProfile")
		if err != nil {
			continue
		}
		if profile, ok := v.Value().(string); ok {
			return profile, true
		}
	}
	return "", false
}

func readOnBattery(conn *dbus.Conn) (bool, bool) {
	v, err := conn.Object(upowerDest, upowerPath).GetProperty(upowerDest + ".OnBattery")
	if err != nil {
		return false, false
	}
	onBattery, ok := v.Value().(bool)
	return onBattery, ok
}

// readBatteryPercent returns the charge of the combined battery, or -1.
func readBatteryPercent(conn *dbus.Conn) float64 {
	obj := conn.Object(upowerDest, upowerDisplayPath)
	present, err := obj.GetProperty(upowerDeviceIface + ".IsPresent")
	if err != nil || present.Value() != true {
		return -1
	}
	kind, err := obj.GetProperty(upowerDeviceIface + ".Type")
	if err != nil || kind.Value() != uint32(upowerTypeBattery) {
		return -1
	}
	percent, err := obj.GetProperty(upowerDeviceIface + ".Percentage")
	if err != nil {
		return -1
	}
	if p, ok := percent.Value().(float64); ok {
		return p
	}
	return -1
}

// watchPortal reads the desktop animation settings and follows their changes.
func watchPortal(m *Monitor, conn *dbus.Conn) {
	obj := conn.Object(portalDest, portalPath)
	settings := map[[2]string]dbus.Variant{}
	apply := func() {
		reduce := false
		if v, ok := settings[gnomeAnimations]; ok {
			if enabled, ok := v.Value().(bool); ok && !enabled {
				reduce = true
			}
		}
		if v, ok := settings[kdeAnimations]; ok {
			if factor, ok := v.Value().(float64); ok && factor == 0 {
				reduce = true
			}
		}
		m.update(func(s *State) {
			s.ReduceMotion = reduce
		})
	}
	for _, key := range [][2]string{gnomeAnimations, kdeAnimations} {
		if v, ok := readPortalSetting(obj, key); ok {
			settings[key] = v
		}
	}
	apply()

	_ = conn.AddMatchSignal(
		dbus.WithMatchObjectPath(portalPath),
		dbus.WithMatchInterface(portalSettings),
		dbus.WithMatchMember("SettingChanged"),
	)
	signals := make(chan *dbus.Signal, 16)
	conn.Signal(signals)
	go func() {
		for sig := range signals {
			if sig.Name != portalSettings+".SettingChanged" || len(sig.Body) != 3 {
				continue
			}
			ns, _ := sig.Body[0].(string)
			key, _ := sig.Body[1].(string)
			value, ok := sig.Body[2].(dbus.Variant)
			k := [2]string{ns, key}
			if !ok || (k != gnomeAnimations && k != kdeAnimations) {
				continue
			}
			settings[k] = value
			apply()
		}
	}()
}

// readPortalSetting reads a setting with ReadOne, falling back to the
// deprecated Read that wraps the value in an extra variant.
func readPortalSetting(obj dbus.BusObject, key [2]string) (dbus.Variant, bool) {
	var v dbus.Variant
	if err := obj.Call(portalSettings+".ReadOne", 0, key[0], key[1]).Store(&v); err == nil {
		return v, true
	}
	if err := obj.Call(portalSettings+".Read", 0, key[0], key[1]).Store(&v); err != nil {
		return dbus.Variant{}, false
	}
	if inner, ok := v.Value().(dbus.Variant); ok {
		return inner, true
	}
	return v, true
}
