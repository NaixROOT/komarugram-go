// SPDX-License-Identifier: Unlicense OR MIT

package block

type Axis int

const (
	AxisHorizontal Axis = 1 << iota
	AxisVertical
)

type Align int

const (
	// AlignBaseline positions blocks based on their baseline, only valid for AxisHorizontal.
	AlignBaseline Align = iota
	// AlignStart positions blocks at the starting edge of the layout's main axis.
	AlignStart
	// AlignMiddle positions blocks in the center along the main axis of the layout.
	AlignMiddle
	// AlignEnd positions blocks at the ending edge of the layout's main axis.
	AlignEnd
)

type Overflow int

const (
	OverflowClip Overflow = 1 << iota
	OverflowWrap
)
