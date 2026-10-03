// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"gio-mw/widget/scroll"

	"gioui.org/io/pointer"
)

// logScroll writes every scroll event the lists receive to the file at
// path, one line each, for measuring what a wheel notch and a touchpad
// gesture send on each platform. The header tells the platform.
func logScroll(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	session := os.Getenv("XDG_SESSION_TYPE")
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		session += " WAYLAND_DISPLAY"
	}
	if os.Getenv("DISPLAY") != "" {
		session += " DISPLAY"
	}
	fmt.Fprintf(f, "# %s/%s session=%q wheel scale %g touchpad scale %g continuous scale %g\n", runtime.GOOS, runtime.GOARCH, session, scroll.WheelScale, scroll.TouchpadScale, scroll.ContinuousScale)
	fmt.Fprintln(f, "# since start (ms), since previous (ms), event time (ms), scroll x, scroll y, modifiers, wheel, continuous, distance (px), applied")
	var mu sync.Mutex
	start := time.Now()
	var last time.Time
	scroll.Trace = func(e pointer.Event, distance float32, precise bool) {
		mu.Lock()
		defer mu.Unlock()
		now := time.Now()
		gap := 0.0
		if !last.IsZero() {
			gap = float64(now.Sub(last).Microseconds()) / 1000
		}
		last = now
		applied := "animated"
		if precise {
			applied = "at once"
		}
		// Each line is written at once, so that nothing is lost when the
		// client is killed.
		fmt.Fprintf(f, "%.3f\t%.3f\t%d\t%g\t%g\t%v\t%v\t%v\t%g\t%s\n", float64(now.Sub(start).Microseconds())/1000, gap, e.Time.Milliseconds(), e.Scroll.X, e.Scroll.Y, e.Modifiers, e.Wheel, e.Continuous, distance, applied)
	}
	return nil
}
