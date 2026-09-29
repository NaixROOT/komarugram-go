// SPDX-License-Identifier: Unlicense OR MIT

package ui

// RecoverFrame leaves a settings subsection whose layout panicked. It runs
// after appwindow has discarded the entire failed frame. Update failures and
// failures outside this boundary keep the window's normal fallback screen.
func (a *App) RecoverFrame() bool {
	p := a.settings
	if p == nil || !p.layingOut {
		return false
	}
	p.layingOut = false
	if p.section == settingsMain {
		return false
	}
	p.section = settingsMain
	// Position alone is not enough: Gio's list retains an unfinished child
	// after a layout panic, and would panic again on the next frame.
	p.scrollPage = scrollPage{}
	p.confirmingLogOut = false
	if p.sessions != nil {
		p.sessions.dialog = modal{}
	}
	p.toast.Show(a.catalog().T("settings.recovered"))
	return true
}
