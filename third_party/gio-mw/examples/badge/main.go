// SPDX-License-Identifier: Unlicense OR MIT

package main

import (
	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/badge"
	"gio-mw/widget/button"
	"gio-mw/widget/overlay"
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

func main() {
	go func() {
		var appWindow app.Window
		appWindow.Option(app.Title("Badge Example"), app.Size(unit.Dp(420), unit.Dp(640)))
		if err := appRun(&appWindow); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	app.Main()
}

var (
	appOverlay          overlay.Overlay
	badgeOne            *badge.Badge
	badgeTwo            *badge.Badge
	badgeThree          *badge.Badge
	badgeFour           *badge.Badge
	badgeFive           *badge.Badge
	toggleBadgeBtn      *button.Button
	mailIcon            = wdk.RequireIconWidget(icons.ContentMail)
	notificationsIcon   = wdk.RequireIconWidget(icons.SocialNotifications)
	accountIcon         = wdk.RequireIconWidget(icons.ActionAccountCircle)
	settingsIcon        = wdk.RequireIconWidget(icons.ActionSettings)
	cropIcon            = wdk.RequireIconWidget(icons.ImageCropSquare)
	mailButton          = button.Text()
	notificationsButton = button.Text()
	accountButton       = button.Text()
	settingsButton      = button.Text()
	cropButton          = button.Text()
)

func appRun(appWindow *app.Window) error {
	var ops op.Ops
	var materialTheme *token.Theme

	badgeOne = &badge.Badge{Visible: true}
	badgeTwo = &badge.Badge{Visible: true}
	badgeThree = &badge.Badge{Visible: true}
	badgeFour = &badge.Badge{Visible: true}
	badgeFive = &badge.Badge{Visible: true}

	toggleBadgeBtn = button.Text()

	for {
		switch windowEvent := appWindow.Event().(type) {
		case app.DestroyEvent:
			return windowEvent.Err

		case app.FrameEvent:
			gtx := app.NewContext(&ops, windowEvent)
			gtx.Values = make(map[string]any)
			if materialTheme == nil {
				scheme := schemes.SchemeBaselineLight()
				materialTheme = defaults.NewTheme(gtx, scheme)
			}
			wdk.InitMaterialThemeInContext(gtx, materialTheme)

			appUpdate(gtx)
			appLayout(gtx)

			windowEvent.Frame(gtx.Ops)
		}
	}
}

func appUpdate(gtx layout.Context) {
	if toggleBadgeBtn.Clicked(gtx) {
		badgeOne.Visible = !badgeOne.Visible
		badgeTwo.Visible = !badgeTwo.Visible
		badgeThree.Visible = !badgeThree.Visible
		badgeFour.Visible = !badgeFour.Visible
		badgeFive.Visible = !badgeFive.Visible
	}
	mailButton.Update(gtx)
	notificationsButton.Update(gtx)
	accountButton.Update(gtx)
	settingsButton.Update(gtx)
	cropButton.Update(gtx)
	appOverlay.Update(gtx)
}

func appLayout(gtx layout.Context) {
	exp.Background(gtx)
	layoutForeground(gtx)
	appOverlay.Layout(gtx)
}

func layoutForeground(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				presentation := wdk.LabelStyle{
					Typestyle: token.TypestyleLabelLarge,
				}
				return wdk.LayoutLabel(gtx, presentation, "Badge Examples")
			}),
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return toggleBadgeBtn.Layout(gtx, "Toggle Badges")
			}),
			block.NewVerticalSpacer(examples.SpacingMedium),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				return block.Line{
					Axis:     block.AxisHorizontal,
					Overflow: block.OverflowClip,
					Expand:   true,
				}.Layout(gtx,
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						iconWidget := badgeOne.OnIconWidget(mailIcon, 0)
						return mailButton.LayoutIconOnly(gtx, "Mail", iconWidget)
					}),
					block.NewHorizontalSpacer(examples.SpacingMedium),
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						iconWidget := badgeTwo.OnIconWidget(notificationsIcon, 1)
						return notificationsButton.LayoutIconOnly(gtx, "Notifications", iconWidget)
					}),
					block.NewHorizontalSpacer(examples.SpacingMedium),
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						iconWidget := badgeThree.OnIconWidget(accountIcon, 10)
						return accountButton.LayoutIconOnly(gtx, "Account", iconWidget)
					}),
					block.NewHorizontalSpacer(examples.SpacingMedium),
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						iconWidget := badgeFour.OnIconWidget(settingsIcon, 102)
						return settingsButton.LayoutIconOnly(gtx, "Settings", iconWidget)
					}),
					block.NewHorizontalSpacer(examples.SpacingMedium),
					block.NewSegment(func(gtx layout.Context) layout.Dimensions {
						iconWidget := badgeFive.OnIconWidget(cropIcon, 1024)
						return cropButton.LayoutIconOnly(gtx, "Crop", iconWidget)
					}),
				)
			}),
		)
	})
}
