// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android

package appearance

import (
	"sync"

	"github.com/godbus/dbus/v5"
)

const (
	portalDest     = "org.freedesktop.portal.Desktop"
	portalPath     = "/org/freedesktop/portal/desktop"
	portalSettings = "org.freedesktop.portal.Settings"

	xfconfDest  = "org.xfce.Xfconf"
	xfconfPath  = "/org/xfce/Xfconf"
	xfconfIface = "org.xfce.Xfconf"
)

// The sources, in order of preference.
var (
	portalScheme = [2]string{"org.freedesktop.appearance", "color-scheme"}
	gnomeScheme  = [2]string{"org.gnome.desktop.interface", "color-scheme"}
	gnomeTheme   = [2]string{"org.gnome.desktop.interface", "gtk-theme"}
	xfceTheme    = [2]string{"xsettings", "/Net/ThemeName"}
)

// sources holds the last value of every source.
type sources struct {
	mu     sync.Mutex
	values map[[2]string]dbus.Variant
}

func (s *sources) put(key [2]string, v dbus.Variant) {
	s.mu.Lock()
	s.values[key] = v
	s.mu.Unlock()
}

// scheme resolves the sources into a scheme.
func (s *sources) scheme() Scheme {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.values[portalScheme].Value().(uint32); ok {
		switch v {
		case 1:
			return Dark
		case 2:
			return Light
		}
	}
	if v, ok := s.values[gnomeScheme].Value().(string); ok {
		switch v {
		case "prefer-dark":
			return Dark
		case "prefer-light":
			return Light
		}
	}
	// XFCE keeps its theme in xfconf; the GNOME key may be stale there.
	for _, key := range [][2]string{xfceTheme, gnomeTheme} {
		if v, ok := s.values[key].Value().(string); ok && v != "" {
			return schemeOfTheme(v)
		}
	}
	return Unknown
}

func startPlatform(m *Monitor) func() {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil
	}
	src := &sources{values: make(map[[2]string]dbus.Variant)}
	portal := conn.Object(portalDest, portalPath)
	for _, key := range [][2]string{portalScheme, gnomeScheme, gnomeTheme} {
		if v, ok := readPortalSetting(portal, key); ok {
			src.put(key, v)
		}
	}
	var theme dbus.Variant
	if err := conn.Object(xfconfDest, xfconfPath).Call(xfconfIface+".GetProperty", 0, xfceTheme[0], xfceTheme[1]).Store(&theme); err == nil {
		src.put(xfceTheme, theme)
	}
	m.set(src.scheme())

	_ = conn.AddMatchSignal(
		dbus.WithMatchObjectPath(portalPath),
		dbus.WithMatchInterface(portalSettings),
		dbus.WithMatchMember("SettingChanged"),
	)
	_ = conn.AddMatchSignal(
		dbus.WithMatchObjectPath(xfconfPath),
		dbus.WithMatchInterface(xfconfIface),
		dbus.WithMatchMember("PropertyChanged"),
	)
	signals := make(chan *dbus.Signal, 16)
	conn.Signal(signals)
	go func() {
		// The channel is closed when the connection is closed.
		for sig := range signals {
			if len(sig.Body) != 3 {
				continue
			}
			ns, _ := sig.Body[0].(string)
			key, _ := sig.Body[1].(string)
			value, ok := sig.Body[2].(dbus.Variant)
			if !ok {
				continue
			}
			k := [2]string{ns, key}
			switch {
			case sig.Name == portalSettings+".SettingChanged" && (k == portalScheme || k == gnomeScheme || k == gnomeTheme),
				sig.Name == xfconfIface+".PropertyChanged" && k == xfceTheme:
				src.put(k, value)
				m.set(src.scheme())
			}
		}
	}()
	return func() { conn.Close() }
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
