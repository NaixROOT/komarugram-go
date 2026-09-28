// SPDX-License-Identifier: Unlicense OR MIT

package router

import (
	"gio-mw/wdk"
	"gio-mw/widget/card"

	"gioui.org/layout"
)

type PageUrl string

type PageWidget interface {
	IsWide() bool
	Update(gtx layout.Context)
	View(gtx layout.Context) layout.Dimensions
}

type Route struct {
	PageUrl     PageUrl
	RedirectTo  PageUrl
	PageIcon    wdk.IconWidget
	PageTitle   string
	PagePreview func(*card.Card) layout.Widget
	PageWidget  func() PageWidget
	// TODO: DataResolvers
}

type Routes []*Route
