// SPDX-License-Identifier: Unlicense OR MIT

package slider

type kind int

const (
	standardKind kind = iota
	centeredKind
	rangeKind
)

func StandardSlider(options []int, value int, onChange func(int)) *Slider {
	numOptions := len(options)
	if numOptions < 3 {
		panic("Slider: must have at least 3 options")
	}
	s := &Slider{
		sKind:       standardKind,
		sOptions:    options,
		sStartValue: value,
		sOnChange:   onChange,
	}
	s.SetValue(value)
	return s
}

func CenteredSlider(options []int, value int, onChange func(int)) *Slider {
	numOptions := len(options)
	if numOptions < 3 {
		panic("Slider: must have at least 3 options")
	}
	if numOptions%2 == 0 {
		panic("Slider: must have an odd number of options")
	}
	s := &Slider{
		sKind:       centeredKind,
		sOptions:    options,
		sStartValue: value,
	}
	s.SetValue(value)
	return s
}

func RangeSlider(options []int, start int, end int) *Slider {
	numOptions := len(options)
	if numOptions < 3 {
		panic("Slider: must have at least 3 options")
	}
	s := &Slider{
		sKind:    rangeKind,
		sOptions: options,
	}
	s.SetRangeValue(start, end)
	return s
}
