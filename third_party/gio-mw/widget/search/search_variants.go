// SPDX-License-Identifier: Unlicense OR MIT

package search

import (
	"gio-mw/wdk"
)

func Bar() *Search {
	return &Search{
		editor: wdk.NewEditor(true, true),
	}
}
