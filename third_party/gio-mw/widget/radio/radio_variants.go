// SPDX-License-Identifier: Unlicense OR MIT

package radio

import (
	"gioui.org/widget"
	"slices"
)

func NewRadios[T comparable](options []T, value T, onChange func(T)) *Radios[T] {
	clickableMap := make(map[T]*widget.Clickable, len(options))
	disabledMap := make(map[T]bool, len(options))
	for _, option := range options {
		clickableMap[option] = &widget.Clickable{}
		disabledMap[option] = false
	}
	if !slices.Contains(options, value) {
		panic("radio.NewRadios: value must be in options")
	}
	return &Radios[T]{
		clickable: clickableMap,
		disabled:  disabledMap,
		options:   options,
		value:     value,
		onChange:  onChange,
	}
}
