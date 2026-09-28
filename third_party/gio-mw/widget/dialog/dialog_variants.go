// SPDX-License-Identifier: Unlicense OR MIT

package dialog

import (
	"gio-mw/widget/button"

	"gioui.org/layout"
)

func Acknowledge(gtx layout.Context) *BasicStyle {
	return &BasicStyle{
		dTheme:        BuildBasicTheme(gtx),
		ConfirmButton: button.Text(),
		ConfirmText:   "OK",
	}
}

func Confirm(gtx layout.Context) *BasicStyle {
	return &BasicStyle{
		dTheme:        BuildBasicTheme(gtx),
		CancelButton:  button.Text(),
		ConfirmButton: button.Text(),
		ConfirmText:   "Confirm",
		CancelText:    "Cancel",
	}
}
