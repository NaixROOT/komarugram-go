// SPDX-License-Identifier: Unlicense OR MIT

package components

import (
	"gio-mw/exp/router"
	"gio-mw/widget/overlay"
	"gio-mw/widget/rail"

	"gioui.org/layout"
)

type AppRail struct {
	appRail   *rail.Rail
	appRouter *router.Router
}

func NewAppRail(appRouter *router.Router) *AppRail {
	currentPageUrl := appRouter.GetCurrentPageUrl()

	var activeRailButton *rail.Button
	var railButtons []*rail.Button
	for _, route := range appRouter.Routes {
		if route.RedirectTo != "" {
			continue
		}
		railButton := &rail.Button{Data: route}
		if route.PageUrl == currentPageUrl {
			activeRailButton = railButton
		}
		railButtons = append(railButtons, railButton)
	}

	return &AppRail{
		appRail: &rail.Rail{
			Active:   activeRailButton,
			Buttons:  railButtons,
			Expanded: true,
		},
		appRouter: appRouter,
	}
}

func (r *AppRail) Open(gtx layout.Context, appOverlay *overlay.Overlay) {
	var railButtons []*rail.ButtonStyle
	for _, route := range r.appRouter.Routes {
		if route.RedirectTo != "" {
			continue
		}
		railButton := &rail.ButtonStyle{
			Icon:  route.PageIcon,
			Label: route.PageTitle,
		}
		railButtons = append(railButtons, railButton)
	}

	appOverlay.Show(r.appRail.AsOverlayItem(railButtons))
}

func (r *AppRail) Update(gtx layout.Context) {
	currentRoute := r.appRouter.GetCurrentRoute()
	activeRailRoute := r.appRail.Active.Data.(*router.Route)
	routeChanged := false
	if currentRoute.PageUrl != activeRailRoute.PageUrl {
		routeChanged = true
	}
	for _, railButton := range r.appRail.Buttons {
		if routeChanged {
			pageRoute := railButton.Data.(*router.Route)
			if pageRoute.PageUrl == currentRoute.PageUrl {
				r.appRail.Active = railButton
			}
		}
		if railButton.Clickable.Clicked(gtx) {
			r.appRail.Active = railButton
			pageRoute := railButton.Data.(*router.Route)
			err := r.appRouter.NavigateTo(pageRoute.PageUrl)
			if err != nil {
				panic(err)
			}
			r.appRail.Close(gtx)
			break
		}
	}
}
