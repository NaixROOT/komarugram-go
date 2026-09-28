// SPDX-License-Identifier: Unlicense OR MIT

// Package powersave watches the system power state and motion preferences,
// so that an application can turn off animations to save battery, the way
// the Telegram clients do.
//
// Support by platform:
//   - Linux: power-profiles-daemon ("power-saver" profile), UPower (battery)
//     and the XDG desktop portal settings (GNOME enable-animations, KDE
//     AnimationDurationFactor), all over D-Bus with change notifications.
//   - Windows: battery saver, battery status and the "Show animations in
//     Windows" setting, polled.
//   - Other platforms: nothing is detected, State stays unknown.
package powersave

import (
	"sync"
)

// State is a snapshot of the system state relevant to power saving.
type State struct {
	// PowerSaver is true when the system power or battery saver is on.
	PowerSaver bool
	// OnBattery is true when the device runs on battery power.
	OnBattery bool
	// BatteryPercent is the battery charge from 0 to 100, or -1 when there
	// is no battery or it is unknown.
	BatteryPercent float64
	// ReduceMotion is true when the user asked the system to turn off or
	// reduce animations.
	ReduceMotion bool
}

// Monitor tracks the system State. Its zero value is not usable, use Start.
type Monitor struct {
	mu       sync.Mutex
	state    State
	onChange func(State)
	stop     func()
}

// Start begins watching the system. onChange, when not nil, is called from
// another goroutine whenever the state changes; a typical implementation
// invalidates the application window. Sources that are unavailable are
// ignored, so Start always succeeds.
func Start(onChange func(State)) *Monitor {
	m := &Monitor{
		state:    State{BatteryPercent: -1},
		onChange: onChange,
	}
	m.stop = startPlatform(m)
	return m
}

// State returns the current state.
func (m *Monitor) State() State {
	if m == nil {
		return State{BatteryPercent: -1}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// Close stops watching the system.
func (m *Monitor) Close() {
	if m != nil && m.stop != nil {
		m.stop()
	}
}

// update applies f to the state and reports a change to onChange.
func (m *Monitor) update(f func(*State)) {
	m.mu.Lock()
	old := m.state
	f(&m.state)
	state := m.state
	m.mu.Unlock()
	if state != old && m.onChange != nil {
		m.onChange(state)
	}
}

// Mode selects when animations are shown.
type Mode int

const (
	// ModeAuto shows animations unless the system is saving power, the
	// battery is low or the user turned animations off system-wide.
	ModeAuto Mode = iota
	// ModeOn always shows animations.
	ModeOn
	// ModeOff never shows animations.
	ModeOff
)

// DefaultLowBattery is the default Policy.LowBattery threshold in percent.
const DefaultLowBattery = 20

// Policy decides whether to animate for a given State.
type Policy struct {
	Mode Mode
	// LowBattery turns animations off in ModeAuto when running on battery
	// at or below this charge in percent. Zero disables the check.
	LowBattery float64
}

// AnimationsEnabled reports whether animations should be shown.
func (p Policy) AnimationsEnabled(s State) bool {
	switch p.Mode {
	case ModeOn:
		return true
	case ModeOff:
		return false
	}
	if s.PowerSaver || s.ReduceMotion {
		return false
	}
	if s.OnBattery && s.BatteryPercent >= 0 && s.BatteryPercent <= p.LowBattery {
		return false
	}
	return true
}
