// SPDX-License-Identifier: Unlicense OR MIT

package token

import (
	"gioui.org/text"
)

type Theme struct {
	Scheme     *Scheme
	TextShaper *text.Shaper
	Typescale  *TypescaleArray
	Widgets    map[string]any
}
