// SPDX-License-Identifier: Unlicense OR MIT

package slider

import (
	"gio-mw/wdk"
	"slices"

	"gioui.org/layout"
)

type Size int

const (
	XSmall Size = iota - 1
	Small
	Medium
	Large
	XLarge
)

type Slider struct {
	Orientation wdk.Axis
	ShowStops   bool
	Size        Size

	sKind       kind
	sOnChange   func(int)
	sOptions    []int
	sEndValue   int
	sEndDrag    wdk.Draggable
	sStartValue int
	sStartDrag  wdk.Draggable
}

func (s *Slider) GetValue() int {
	return s.sStartValue
}

func (s *Slider) SetValue(value int) {
	if !slices.Contains(s.sOptions, value) {
		panic("Slider: value must be in options")
	}
	s.sStartValue = value
}

func (s *Slider) GetRangeValue() (start int, end int) {
	return s.sStartValue, s.sEndValue
}

func (s *Slider) SetRangeValue(start int, end int) {
	if start >= end {
		panic("Slider: start value must be less than end value")
	}
	if !slices.Contains(s.sOptions, start) {
		panic("Slider: start value must be in options")
	}
	if !slices.Contains(s.sOptions, end) {
		panic("Slider: end value must be in options")
	}
	s.sStartValue = start
	s.sEndValue = end
}

func (s *Slider) Layout(gtx layout.Context) layout.Dimensions {
	style := widgetStyle{
		Slider: s,
		Theme:  BuildTheme(gtx),
	}
	return style.layout(gtx)
}
