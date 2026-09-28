// SPDX-License-Identifier: Unlicense OR MIT

package input

import (
	"gio-mw/wdk"
)

func FilledTextInput() *Input {
	return &Input{
		Editor: wdk.NewEditor(true, true),
	}
}

func FilledTextArea() *Input {
	return &Input{
		Editor: wdk.NewEditor(false, false),
	}
}
