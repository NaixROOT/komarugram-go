// SPDX-License-Identifier: Unlicense OR MIT

package toggle

import (
	"gioui.org/widget"
	"slices"
)

func NewToggle[T comparable](options []T, values []T, onChange func([]T)) *Toggle[T] {
	clickableMap := make(map[T]*widget.Clickable, len(options))
	disabledMap := make(map[T]bool, len(options))
	for _, option := range options {
		clickableMap[option] = &widget.Clickable{}
		disabledMap[option] = false
	}
	for _, value := range values {
		if !slices.Contains(options, value) {
			panic("toggle.NewToggle: value must be in options")
		}
	}
	return &Toggle[T]{
		clickable: clickableMap,
		disabled:  disabledMap,
		options:   options,
		values:    values,
		onChange:  onChange,
	}
}
