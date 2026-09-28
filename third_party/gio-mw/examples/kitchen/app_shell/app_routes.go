// SPDX-License-Identifier: Unlicense OR MIT

package app_shell

import (
	"gio-mw/examples/kitchen/pages/app_theme"
	"gio-mw/examples/kitchen/pages/buttons"
	"gio-mw/examples/kitchen/pages/cards"
	"gio-mw/examples/kitchen/pages/checkboxes"
	"gio-mw/examples/kitchen/pages/colors"
	"gio-mw/examples/kitchen/pages/dialogs"
	"gio-mw/examples/kitchen/pages/indicators"
	"gio-mw/examples/kitchen/pages/labels"
	"gio-mw/examples/kitchen/pages/radios"
	"gio-mw/examples/kitchen/pages/search_fields"
	"gio-mw/examples/kitchen/pages/sliders"
	"gio-mw/examples/kitchen/pages/text_fields"
	"gio-mw/examples/kitchen/pages/toggle"
	"gio-mw/exp/router"
	"gio-mw/wdk"

	"golang.org/x/exp/shiny/materialdesign/icons"
)

var appRoutes = router.Routes{
	{
		PageUrl:    "",
		RedirectTo: "app_theme",
	},
	{
		PageUrl:    "app_theme",
		PageTitle:  "App Theme",
		PageWidget: app_theme.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.ImageStyle),
	},
	{
		PageUrl:     "buttons",
		PageTitle:   "Buttons",
		PagePreview: buttons.NewPreview,
		PageWidget:  buttons.NewPage,
		PageIcon:    wdk.RequireIconWidget(icons.ContentFlag),
	},
	{
		PageUrl:    "cards",
		PageTitle:  "Cards",
		PageWidget: cards.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.ActionCreditCard),
	},
	{
		PageUrl:     "checkboxes",
		PageTitle:   "Checkboxes",
		PagePreview: checkboxes.NewPreview,
		PageWidget:  checkboxes.NewPage,
		PageIcon:    wdk.RequireIconWidget(icons.ToggleCheckBox),
	},
	{
		PageUrl:    "colors",
		PageTitle:  "Colors",
		PageWidget: colors.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.EditorBorderColor),
	},
	{
		PageUrl:     "dialogs",
		PageTitle:   "Dialogs",
		PagePreview: dialogs.NewPreview,
		PageWidget:  dialogs.NewPage,
		PageIcon:    wdk.RequireIconWidget(icons.ActionOpenInNew),
	},
	{
		PageUrl:    "indicators",
		PageTitle:  "Indicators",
		PageWidget: indicators.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.NavigationRefresh),
	},
	{
		PageUrl:    "labels",
		PageTitle:  "Labels",
		PageWidget: labels.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.EditorFormatBold),
	},
	{
		PageUrl:     "radios",
		PageTitle:   "Radios",
		PagePreview: radios.NewPreview,
		PageWidget:  radios.NewPage,
		PageIcon:    wdk.RequireIconWidget(icons.ToggleRadioButtonChecked),
	},
	{
		PageUrl:    "search_fields",
		PageTitle:  "Search fields",
		PageWidget: search_fields.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.ActionPageview),
	},
	{
		PageUrl:    "sliders",
		PageTitle:  "Sliders",
		PageWidget: sliders.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.AVVolumeUp),
	},
	{
		PageUrl:    "text_fields",
		PageTitle:  "Text fields",
		PageWidget: text_fields.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.EditorShortText),
	},
	{
		PageUrl:    "toggles",
		PageTitle:  "Toggles",
		PageWidget: toggle.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.ToggleStar),
	},
}
