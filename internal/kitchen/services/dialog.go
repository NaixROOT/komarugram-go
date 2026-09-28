// SPDX-License-Identifier: Unlicense OR MIT

package services

import "gio-mw/widget/dialog"

var (
	appDialogSingleton AppDialog
)

type AppDialog interface {
	ShowBasicDialog(basicDialog *dialog.BasicStyle)
	HideBasicDialog(basicDialog *dialog.BasicStyle)
}

func GetDialogService() AppDialog {
	return appDialogSingleton
}

func SetDialogService(appDialog AppDialog) {
	appDialogSingleton = appDialog
}
