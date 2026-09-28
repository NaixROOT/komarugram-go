// SPDX-License-Identifier: Unlicense OR MIT

package wdk

import (
	"gio-mw/token"

	"gioui.org/layout"
)

type IconWidget func(gtx layout.Context, foreground token.MatColor) layout.Dimensions
