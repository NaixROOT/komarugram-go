// SPDX-License-Identifier: Unlicense OR MIT

package powersave

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	user32                   = syscall.NewLazyDLL("user32.dll")
	procGetSystemPowerStatus = kernel32.NewProc("GetSystemPowerStatus")
	procSystemParametersInfo = user32.NewProc("SystemParametersInfoW")
)

// systemPowerStatus is SYSTEM_POWER_STATUS.
type systemPowerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

const (
	batteryFlagNoBattery    = 128
	batteryPercentUnknown   = 255
	spiGetClientAreaAnimate = 0x1042 // SPI_GETCLIENTAREAANIMATION
	pollInterval            = 10 * time.Second
)

func startPlatform(m *Monitor) func() {
	poll := func() {
		var s systemPowerStatus
		powerOK, _, _ := procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&s)))
		var animate int32 = 1
		animOK, _, _ := procSystemParametersInfo.Call(spiGetClientAreaAnimate, 0, uintptr(unsafe.Pointer(&animate)), 0)
		m.update(func(st *State) {
			if powerOK != 0 {
				st.PowerSaver = s.SystemStatusFlag == 1
				st.OnBattery = s.ACLineStatus == 0
				st.BatteryPercent = -1
				if s.BatteryFlag&batteryFlagNoBattery == 0 && s.BatteryLifePercent != batteryPercentUnknown {
					st.BatteryPercent = float64(s.BatteryLifePercent)
				}
			}
			if animOK != 0 {
				st.ReduceMotion = animate == 0
			}
		})
	}
	poll()
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				poll()
			}
		}
	}()
	var once bool
	return func() {
		if !once {
			once = true
			close(done)
		}
	}
}
