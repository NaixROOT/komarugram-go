// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"testing"
	"time"

	"gioui.org/app"
	"gioui.org/layout"

	"komarugram/internal/appwindow"
	"komarugram/internal/messenger/mockstore"
	"komarugram/internal/motion"
)

func TestSettingsLayoutPanicRecovery(t *testing.T) {
	w := &appwindow.Window{Window: new(app.Window), Motion: motion.New(func() {})}
	defer w.Motion.Close()
	a := New(w, mockstore.New(time.Now(), 0), Services{})
	defer a.Close()
	a.section = section{kind: sectionSettings}
	h := &focusHarness{draw: func(gtx layout.Context) { a.Update(gtx); a.Layout(gtx) }}
	h.frame()
	if a.RecoverFrame() {
		t.Fatal("successful layout must not trigger recovery")
	}
	// Recreate the radio contract violation inside the real settings list.
	// Keep it broken: recovery must leave the section, not retry its layout.
	c := a.settings.decoders.stickers
	c.options = append(c.options, "invalid")
	for range 2 {
		a.settings.section = settingsIntegrations
		var failure any
		func() {
			defer func() { failure = recover() }()
			h.frame()
		}()
		if failure == nil {
			t.Fatal("fault injection did not panic")
		}
		if !a.RecoverFrame() || a.settings.section != settingsMain {
			t.Fatal("failed subsection did not return to settings")
		}
		// Gio's input router consumes the recovered frames too. Reusing the
		// unfinished list or any of the failed frame's macros panics here.
		h.frame()
		h.frame()
		if a.settings.toast.Text() == "" {
			t.Fatal("recovery did not explain the navigation")
		}
		// Navigate through actual click handling after recovery.
		a.settings.items[settingsAppearance].click.Click()
		h.frame()
		if a.settings.section != settingsAppearance {
			t.Fatal("settings stopped handling input after recovery")
		}
		a.settings.open()
		h.frame()
	}
	// Never claim to recover a failure in the main page, or elsewhere.
	a.settings.layingOut = true
	if a.RecoverFrame() || a.RecoverFrame() {
		t.Fatal("main settings failure must use the window fallback")
	}
}
