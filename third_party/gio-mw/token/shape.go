// SPDX-License-Identifier: Unlicense OR MIT

package token

import (
	"gioui.org/unit"
)

type CornerKind int

const (
	CornerKindDefault CornerKind = iota
	CornerKindChamfer
	CornerKindRound
)

type CornerShape struct {
	Kind        CornerKind
	Size        unit.Dp
	AdaptToSize bool
}

type CornerShapes struct {
	TopStart    CornerShape
	TopEnd      CornerShape
	BottomStart CornerShape
	BottomEnd   CornerShape
}

func UniformCornerShapes(corner CornerShape) CornerShapes {
	return CornerShapes{
		BottomEnd:   corner,
		BottomStart: corner,
		TopEnd:      corner,
		TopStart:    corner,
	}
}
