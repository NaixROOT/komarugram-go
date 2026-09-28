// SPDX-License-Identifier: Unlicense OR MIT

package search_fields

import (
	"fmt"
	"komarugram/internal/kitchen/services"

	"gio-mw/exp"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/search"

	"gioui.org/layout"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type Page struct {
	toggleLeadingButton *button.Button
	searchBar           *search.Search
	searchMenuBar       *search.Search
	currentIcon         string
}

var (
	menuIcon   wdk.IconWidget
	searchIcon wdk.IconWidget
)

func NewPage() router.PageWidget {
	menuIcon = wdk.RequireIconWidget(icons.NavigationMenu)
	searchIcon = wdk.RequireIconWidget(icons.ActionSearch)

	searchBar := search.Bar()
	searchBar.LeadingIcon.Icon = searchIcon
	searchBar.TrailingIcon.Icon = wdk.RequireIconWidget(icons.ContentClear)
	searchBar.TrailingIcon.Label = "Clear"
	searchBar.SupportingText = "Supporting text"

	menuBar := search.Bar()
	menuBar.LeadingIcon.Icon = menuIcon
	menuBar.LeadingIcon.Label = "Open Menu"
	menuBar.SupportingText = "Search with menu button"

	return &Page{
		currentIcon:         "searchIcon",
		toggleLeadingButton: button.Text(),
		searchBar:           searchBar,
		searchMenuBar:       menuBar,
	}
}

func (p *Page) IsWide() bool {
	return false
}

func (p *Page) Update(gtx layout.Context) {
	notificationService := services.GetNotificationService()
	if p.toggleLeadingButton.Clicked(gtx) {
		switch p.currentIcon {
		case "searchIcon":
			p.searchBar.LeadingIcon.Icon = menuIcon
			p.searchBar.LeadingIcon.Label = "Menu"
			p.currentIcon = "menuIcon"
		case "menuIcon":
			p.searchBar.LeadingIcon.Icon = searchIcon
			p.searchBar.LeadingIcon.Label = ""
			p.currentIcon = "searchIcon"
		}
	}
	if p.searchBar.LeadingIcon.Clickable.Clicked(gtx) {
		notificationService.ShowNotification("Menu clicked")
	}
	if p.searchBar.Submitted(gtx) {
		searchQuery := p.searchBar.GetText()
		notificationService.ShowNotification(fmt.Sprintf("Searching for '%s'", searchQuery))
	}
	if p.searchBar.TrailingIcon.Clickable.Clicked(gtx) {
		p.searchBar.ClearText()
	}
	if p.searchMenuBar.LeadingIcon.Clickable.Clicked(gtx) {
		searchQuery := p.searchMenuBar.GetText()
		notificationService.ShowNotification(fmt.Sprintf("Opening menu with query '%s'", searchQuery))
	}
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Search fields"
				return exp.HeadlineL(gtx, txt)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				txt := "Let people enter a keyword or phrase to get relevant information"
				return exp.BodyL(gtx, txt)
			}),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.sectionConfig),
			block.NewVerticalSpacer(examples.SpacingSmall),
			block.NewSegment(p.sectionExamples),
		)
	})
}

func (p *Page) sectionConfig(gtx layout.Context) layout.Dimensions {
	return block.Line{
		Axis:     block.AxisHorizontal,
		Overflow: block.OverflowWrap,
		Expand:   true,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return p.toggleLeadingButton.Layout(gtx, "Toggle Leading Button")
		}),
		block.NewVerticalSpacer(examples.SpacingSmall),
	)
}

func (p *Page) sectionExamples(gtx layout.Context) layout.Dimensions {
	wdk.EnforceMax(&gtx, 480, gtx.Constraints.Max.Y)
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(p.searchBar.Layout),
		block.NewVerticalSpacer(examples.SpacingMedium),
		block.NewSegment(p.searchMenuBar.Layout),
		block.NewVerticalSpacer(examples.SpacingMedium),
	)
}
