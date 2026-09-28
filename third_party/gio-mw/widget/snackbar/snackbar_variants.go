// SPDX-License-Identifier: Unlicense OR MIT

package snackbar

func Plain(supportingText string) *Style {
	return &Style{
		supportingText: supportingText,
	}
}
