// SPDX-License-Identifier: Unlicense OR MIT

package services

import (
	"gio-mw/exp/router"
)

var (
	appNavigationSingleton AppNavigation
)

type AppNavigation interface {
	GetRoutesWithPreview() router.Routes
	GetActiveRoute() *router.Route
	NavigateTo(url router.PageUrl) error
}

func GetNavigationService() AppNavigation {
	return appNavigationSingleton
}

func SetNavigationService(appNavigation AppNavigation) {
	appNavigationSingleton = appNavigation
}
