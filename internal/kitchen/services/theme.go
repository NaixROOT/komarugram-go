// SPDX-License-Identifier: Unlicense OR MIT

package services

import (
	"gio-mw/token"

	"gioui.org/layout"
)

var (
	appThemesSingleton AppThemes
)

type AppThemes interface {
	SetTheme(gtx layout.Context, theme *token.Theme)
}

func GetThemeService() AppThemes {
	return appThemesSingleton
}

func SetThemeService(appThemes AppThemes) {
	appThemesSingleton = appThemes
}
