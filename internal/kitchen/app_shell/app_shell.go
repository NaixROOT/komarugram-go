// SPDX-License-Identifier: Unlicense OR MIT

package app_shell

import (
	"komarugram/internal/kitchen/app_shell/components"
	"komarugram/internal/kitchen/services"
	"time"

	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/exp"
	"gio-mw/exp/router"
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/dialog"
	"gio-mw/widget/overlay"
	"gio-mw/widget/scroll"
	"gio-mw/widget/snackbar"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/widget"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var (
	menuIcon = wdk.RequireIconWidget(icons.NavigationMenu)
)

type AppShell struct {
	AppTheme          *token.Theme
	appOverlay        *overlay.Overlay
	appRail           *components.AppRail
	appRailToggle     *button.Button
	appRouter         *router.Router
	appWindow         *app.Window
	appTabsScrollList *widget.List
	appViewScrollList *scroll.List
}

func NewAppShell(appWindow *app.Window) *AppShell {
	appRouter := router.NewRouter(appWindow, appRoutes)
	appRail := components.NewAppRail(appRouter)

	appShell := &AppShell{
		appOverlay:    &overlay.Overlay{},
		appRail:       appRail,
		appRailToggle: button.Text(),
		appRouter:     appRouter,
		appWindow:     appWindow,
		appTabsScrollList: &widget.List{
			Scrollbar: widget.Scrollbar{},
			List:      layout.List{Axis: layout.Horizontal},
		},
		appViewScrollList: &scroll.List{
			List: layout.List{Axis: layout.Vertical},
		},
	}
	services.SetDialogService(appShell)
	services.SetNotificationService(appShell)
	services.SetThemeService(appShell)
	services.SetNavigationService(appShell)
	return appShell
}

func (shell *AppShell) ShowBasicDialog(basicDialog *dialog.BasicStyle) {
	shell.appOverlay.Show(basicDialog.AsOverlayItem())
}

func (shell *AppShell) HideBasicDialog(basicDialog *dialog.BasicStyle) {
	shell.appOverlay.ClearItem(basicDialog.GetLayoutItemId())
}

func (shell *AppShell) ShowNotification(msg string) {
	snackbarStyle := snackbar.Plain(msg)
	shell.appOverlay.Show(overlay.NewItem(snackbarStyle.Layout, block.GravityBottomCenter).WithDuration(time.Second))
}

func (shell *AppShell) GetRoutesWithPreview() router.Routes {
	var routesWithPreview router.Routes
	for _, route := range shell.appRouter.Routes {
		if route.PagePreview != nil {
			routesWithPreview = append(routesWithPreview, route)
		}
	}
	return routesWithPreview
}

func (shell *AppShell) GetActiveRoute() *router.Route {
	return shell.appRouter.GetCurrentRoute()
}

func (shell *AppShell) NavigateTo(url router.PageUrl) error {
	return shell.appRouter.NavigateTo(url)
}

// Theme returns the application theme, creating the default one on first use.
func (shell *AppShell) Theme(gtx layout.Context) *token.Theme {
	if shell.AppTheme == nil {
		shell.AppTheme = defaults.NewTheme(gtx, schemes.SchemeBaselineLight())
	}
	return shell.AppTheme
}

func (shell *AppShell) SetTheme(gtx layout.Context, theme *token.Theme) {
	shell.AppTheme = theme
	wdk.InitMaterialThemeInContext(gtx, theme)
	shell.appRouter.ClearCachedPage()
}

func (shell *AppShell) Layout(gtx layout.Context) {
	exp.Background(gtx)

	block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
		Expand:   true,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = 0
			return block.UniformPadding(16).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return block.Line{
					Axis:     block.AxisHorizontal,
					Overflow: block.OverflowClip,
				}.Layout(gtx,
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						return shell.appRailToggle.LayoutIconOnly(gtx, "Menu", menuIcon)
					}),
					block.NewHorizontalSpacer(16),
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						currentRoute := shell.appRouter.GetCurrentRoute()
						return currentRoute.PageIcon(gtx, shell.AppTheme.Scheme.Surface.OnColor)
					}).AlignMiddle(),
				)
			})
		}),
		block.NewFlexSegment(func(gtx layout.Context) layout.Dimensions {
			availableHeight := gtx.Constraints.Max.Y
			return shell.appViewScrollList.Layout(gtx, 1, func(gtx layout.Context, index int) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				segmentWidget := func(gtx layout.Context) layout.Dimensions {
					pageWidget := *(shell.appRouter.GetCurrentRoutePageWidget())
					pageSizeBreakpoint := gtx.Dp(960)
					if !pageWidget.IsWide() {
						if gtx.Constraints.Max.X > pageSizeBreakpoint {
							gtx.Constraints.Max.X = pageSizeBreakpoint
						}
						if gtx.Constraints.Min.X > gtx.Constraints.Max.X {
							gtx.Constraints.Min.X = gtx.Constraints.Max.X
						}
					}
					gtx.Constraints.Min.Y = availableHeight
					pageWidget.Update(gtx)
					return pageWidget.View(gtx)
				}
				return block.Line{
					Axis:     block.AxisVertical,
					Overflow: block.OverflowClip,
				}.Layout(gtx,
					block.Segment{CrossAlign: block.AlignMiddle, Widget: segmentWidget},
				)
			})
		}),
	)
	shell.appOverlay.Layout(gtx)
}

func (shell *AppShell) Update(gtx layout.Context) {
	if shell.appRailToggle.Clicked(gtx) {
		shell.appRail.Open(gtx, shell.appOverlay)
	}
	shell.appRail.Update(gtx)
	shell.appOverlay.Update(gtx)
}
