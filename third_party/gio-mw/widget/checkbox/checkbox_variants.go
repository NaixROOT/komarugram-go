// SPDX-License-Identifier: Unlicense OR MIT

package checkbox

import "gioui.org/widget"

func NewCheckboxes[T comparable](options []T, values []T, onChange func([]T)) *Checkboxes[T] {
	clickableMap := make(map[T]*widget.Clickable)
	disabledMap := make(map[T]bool)
	errorMap := make(map[T]bool)
	for _, option := range options {
		clickableMap[option] = &widget.Clickable{}
		disabledMap[option] = false
		errorMap[option] = false
	}
	return &Checkboxes[T]{
		parent:    &widget.Clickable{},
		clickable: clickableMap,
		disabled:  disabledMap,
		error:     errorMap,
		onChange:  onChange,
		options:   options,
		values:    values,
	}
}
