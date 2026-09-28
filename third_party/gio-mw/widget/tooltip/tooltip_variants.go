// SPDX-License-Identifier: Unlicense OR MIT

package tooltip

func PlainTooltip(txt string) *Tooltip {
	return &Tooltip{
		SupportingText: txt,
	}
}
