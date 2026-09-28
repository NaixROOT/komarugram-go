// SPDX-License-Identifier: Unlicense OR MIT

package app_shell

import (
	"komarugram/internal/kitchen/pages/app_theme"
	"komarugram/internal/kitchen/pages/buttons"
	"komarugram/internal/kitchen/pages/cards"
	"komarugram/internal/kitchen/pages/checkboxes"
	"komarugram/internal/kitchen/pages/colors"
	"komarugram/internal/kitchen/pages/dialogs"
	"komarugram/internal/kitchen/pages/external"
	"komarugram/internal/kitchen/pages/indicators"
	"komarugram/internal/kitchen/pages/labels"
	"komarugram/internal/kitchen/pages/lottie"
	"komarugram/internal/kitchen/pages/miniapp"
	"komarugram/internal/kitchen/pages/radios"
	"komarugram/internal/kitchen/pages/search_fields"
	"komarugram/internal/kitchen/pages/sliders"
	"komarugram/internal/kitchen/pages/sticker"
	"komarugram/internal/kitchen/pages/text_fields"
	"komarugram/internal/kitchen/pages/toggle"
	"komarugram/internal/kitchen/pages/video"
	"komarugram/internal/kitchen/pages/video_grid"

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
		PageUrl:    "external_player",
		PageTitle:  "External player",
		PageWidget: external.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.AVFeaturedVideo),
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
		PageUrl:    "lottie",
		PageTitle:  "Lottie",
		PageWidget: lottie.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.ImageFilterFrames),
	},
	{
		PageUrl:    "mini_apps",
		PageTitle:  "Mini Apps",
		PageWidget: miniapp.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.ActionExtension),
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
		PageUrl:    "stickers",
		PageTitle:  "Video stickers",
		PageWidget: sticker.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.ImagePhotoFilter),
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
	{
		PageUrl:    "video",
		PageTitle:  "Video",
		PageWidget: video.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.AVMovie),
	},
	{
		PageUrl:    "video_grid",
		PageTitle:  "Video grid",
		PageWidget: video_grid.NewPage,
		PageIcon:   wdk.RequireIconWidget(icons.ImageGridOn),
	},
}
