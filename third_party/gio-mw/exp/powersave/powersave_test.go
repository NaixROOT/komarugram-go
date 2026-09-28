// SPDX-License-Identifier: Unlicense OR MIT

package powersave

import "testing"

func TestPolicy(t *testing.T) {
	auto := Policy{Mode: ModeAuto, LowBattery: 20}
	tests := []struct {
		name   string
		policy Policy
		state  State
		want   bool
	}{
		{"auto, idle desktop", auto, State{BatteryPercent: -1}, true},
		{"auto, power saver", auto, State{PowerSaver: true, BatteryPercent: -1}, false},
		{"auto, reduce motion", auto, State{ReduceMotion: true, BatteryPercent: -1}, false},
		{"auto, battery above threshold", auto, State{OnBattery: true, BatteryPercent: 21}, true},
		{"auto, low battery", auto, State{OnBattery: true, BatteryPercent: 20}, false},
		{"auto, low battery but charging", auto, State{BatteryPercent: 5}, true},
		{"auto, unknown battery level", auto, State{OnBattery: true, BatteryPercent: -1}, true},
		{"auto, threshold disabled", Policy{Mode: ModeAuto}, State{OnBattery: true, BatteryPercent: 5}, true},
		{"on overrides power saver", Policy{Mode: ModeOn}, State{PowerSaver: true}, true},
		{"off", Policy{Mode: ModeOff}, State{BatteryPercent: -1}, false},
	}
	for _, tt := range tests {
		if got := tt.policy.AnimationsEnabled(tt.state); got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}
