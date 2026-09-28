// SPDX-License-Identifier: Unlicense OR MIT

package router

import (
	"fmt"
	"gioui.org/app"
	"log"
)

var currentRouter *Router

type Router struct {
	appWindow      *app.Window
	currentPageUrl PageUrl
	initialized    bool
	Routes         Routes
	cachedPage     *PageWidget
}

func NewRouter(appWindow *app.Window, appRoutes Routes) *Router {
	if currentRouter != nil {
		panic(fmt.Errorf("the Router is already initialized"))
	}
	currentRouter = &Router{appWindow: appWindow, Routes: appRoutes}
	currentRouter.initCurrentRoute()
	return currentRouter
}

func (r *Router) initCurrentRoute() {
	if !r.initialized {
		initialRoute := r.GetRouteByPageUrl(r.currentPageUrl)
		if initialRoute.RedirectTo != "" {
			r.initialized = true
			err := r.NavigateTo(initialRoute.RedirectTo)
			if err != nil {
				panic(fmt.Errorf("invalid RedirectTo for initial route: %w", err))
			}
			return
		}
		if initialRoute.PageWidget == nil {
			panic(fmt.Errorf("missing PageWidget for initial route"))
		}
	}
}

func (r *Router) GetRouteByPageUrl(url PageUrl) *Route {
	for _, route := range r.Routes {
		if route.PageUrl == url {
			return route
		}
	}
	if url == "" {
		return r.Routes[0]
	}
	panic(fmt.Errorf("no such URL: %v", url))
}

func (r *Router) NavigateTo(url PageUrl) error {
	nextRoute := r.GetRouteByPageUrl(url)
	if nextRoute != nil {
		if nextRoute.RedirectTo != "" {
			nextRoute = r.GetRouteByPageUrl(nextRoute.RedirectTo)
		}
	}
	if nextRoute != nil {
		r.cachedPage = nil
		r.currentPageUrl = url
		r.appWindow.Option(app.Title("M3 Kitchen: " + nextRoute.PageTitle))
		return nil
	} else {
		return fmt.Errorf("no such URL: %v", url)
	}
}

func (r *Router) GetCurrentPageUrl() PageUrl {
	return r.currentPageUrl
}

func (r *Router) GetCurrentRoute() *Route {
	route := r.GetRouteByPageUrl(r.currentPageUrl)
	if route.PageWidget == nil {
		log.Fatal("invalid route")
	}
	return route
}

func (r *Router) GetCurrentRoutePageWidget() *PageWidget {
	if r.cachedPage != nil {
		return r.cachedPage
	}
	pageWidget := r.GetCurrentRoute().PageWidget()
	r.cachedPage = &pageWidget
	return r.cachedPage
}

func (r *Router) ClearCachedPage() {
	r.cachedPage = nil
}
