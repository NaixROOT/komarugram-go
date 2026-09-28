// SPDX-License-Identifier: Unlicense OR MIT

package block

import (
	"fmt"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"image"
	"log"
)

// FillDirection represents the fill behavior of a container.
type FillDirection uint8

const (
	FillNone FillDirection = iota
	FillHorizontal
	FillVertical
	FillBoth
)

// Gravity represents the position inside a container, respecting the locale.Direction.
type Gravity int

const (
	GravityTopStart Gravity = iota
	GravityTopCenter
	GravityTopEnd

	GravityMiddleStart
	GravityMiddleCenter
	GravityMiddleEnd

	GravityBottomStart
	GravityBottomCenter
	GravityBottomEnd
)

func (g Gravity) String() string {
	switch g {
	case GravityTopStart:
		return "GravityTopStart"
	case GravityTopCenter:
		return "GravityTopCenter"
	case GravityTopEnd:
		return "GravityTopEnd"
	case GravityMiddleStart:
		return "GravityMiddleStart"
	case GravityMiddleCenter:
		return "GravityMiddleCenter"
	case GravityMiddleEnd:
		return "GravityMiddleEnd"
	case GravityBottomStart:
		return "GravityBottomStart"
	case GravityBottomCenter:
		return "GravityBottomCenter"
	case GravityBottomEnd:
		return "GravityBottomEnd"
	}
	return "unknown"
}

// Container represents a layout container that manages child widget positioning and sizing.
type Container struct {
	// MinSize specifies the minimum size of the container.
	MinSize image.Point
	// MaxSize specifies the maximum size of the container. Zero means unlimited.
	MaxSize image.Point
	// FillDirection specifies which direction(s) the container should fill.
	FillDirection FillDirection
	// Gravity determines the position of the child widget within the container.
	Gravity Gravity
}

func (c Container) Layout(gtx layout.Context, cWidget layout.Widget) layout.Dimensions {
	if err := c.validateConstraints(gtx); err != nil {
		log.Printf("Invalid Container constraints: %v", err)
		return layout.Dimensions{}
	}
	gtx = c.updateConstraints(gtx)

	widgetDimensions, _ := c.layoutWidget(gtx, cWidget)
	return widgetDimensions
}

func (c Container) validateConstraints(gtx layout.Context) error {
	if gtx.Constraints.Max.X == 0 {
		return fmt.Errorf("no horizontal space available")
	}
	if gtx.Constraints.Max.Y == 0 {
		return fmt.Errorf("no vertical space available")
	}
	if c.MinSize.X < 0 || c.MinSize.Y < 0 {
		return fmt.Errorf("MinSize %v < 0", c.MinSize)
	}
	if c.MaxSize.X < 0 || c.MaxSize.Y < 0 {
		return fmt.Errorf("MaxSize %v < 0", c.MaxSize)
	}
	if c.MinSize.X > 0 {
		if gtx.Constraints.Max.X < c.MinSize.X {
			return fmt.Errorf("Gtx.Max.X %v < MinSize.X %v", gtx.Constraints.Max.X, c.MinSize.X)
		}
	}
	if c.MinSize.Y > 0 {
		if gtx.Constraints.Max.Y < c.MinSize.Y {
			return fmt.Errorf("Gtx.Max.Y %v < MinSize.Y %v", gtx.Constraints.Max.Y, c.MinSize.Y)
		}
	}
	if c.MaxSize.X != 0 {
		if c.MinSize.X > c.MaxSize.X {
			return fmt.Errorf("MinSize.X %v > MaxSize.X %v", c.MinSize.X, c.MaxSize.X)
		}
	}
	if c.MaxSize.Y != 0 {
		if c.MinSize.Y > c.MaxSize.Y {
			return fmt.Errorf("invalid MaxSize.Y: %v > %v", c.MinSize.Y, c.MaxSize.Y)
		}
	}
	if c.FillDirection == FillHorizontal || c.FillDirection == FillBoth {
		if c.MinSize.X != 0 {
			return fmt.Errorf("FillX and MinSize.X are mutually exclusive")
		}
		if c.MaxSize.X != 0 {
			return fmt.Errorf("FillX and MaxSize.X are mutually exclusive")
		}
	}
	if c.FillDirection == FillVertical || c.FillDirection == FillBoth {
		if c.MinSize.Y != 0 {
			return fmt.Errorf("FillY and MinSize.Y are mutually exclusive")
		}
		if c.MaxSize.Y != 0 {
			return fmt.Errorf("FillY and MaxSize.Y are mutually exclusive")
		}
	}
	return nil
}

func (c Container) updateConstraints(gtx layout.Context) layout.Context {
	if c.MinSize.X != 0 && c.MinSize.X > gtx.Constraints.Min.X {
		gtx.Constraints.Min.X = c.MinSize.X
	}
	if c.MinSize.Y != 0 && c.MinSize.Y > gtx.Constraints.Min.Y {
		gtx.Constraints.Min.Y = c.MinSize.Y
	}
	if c.MaxSize.X != 0 {
		if gtx.Constraints.Max.X > c.MaxSize.X {
			gtx.Constraints.Max.X = c.MaxSize.X
		}
	}
	if c.MaxSize.Y != 0 {
		if gtx.Constraints.Max.Y > c.MaxSize.Y {
			gtx.Constraints.Max.Y = c.MaxSize.Y
		}
	}
	if c.FillDirection == FillHorizontal || c.FillDirection == FillBoth {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	if c.FillDirection == FillVertical || c.FillDirection == FillBoth {
		gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	}
	return gtx
}

func (c Container) getChildBaseline(childDimensions layout.Dimensions, availableHeight int) int {
	switch c.Gravity {
	case GravityTopStart, GravityTopCenter, GravityTopEnd:
		return availableHeight - childDimensions.Size.Y + childDimensions.Baseline
	case GravityMiddleStart, GravityMiddleCenter, GravityMiddleEnd:
		return (availableHeight-childDimensions.Size.Y)/2 + childDimensions.Baseline
	case GravityBottomStart, GravityBottomCenter, GravityBottomEnd:
		return childDimensions.Baseline
	default:
		return 0
	}
}

func (c Container) clipOverflow(gtx layout.Context, callOp op.CallOp, dims layout.Dimensions) (op.CallOp, layout.Dimensions) {
	clipMax := image.Point{}
	if c.MaxSize.X != 0 && dims.Size.X > c.MaxSize.X {
		clipMax.X = c.MaxSize.X
	}
	if c.MaxSize.Y != 0 && dims.Size.Y > c.MaxSize.Y {
		clipMax.Y = c.MaxSize.Y
	}
	if clipMax.X > 0 || clipMax.Y > 0 {
		if clipMax.X == 0 {
			clipMax.X = dims.Size.X
		}
		if clipMax.Y == 0 {
			clipMax.Y = dims.Size.Y
		}
		macroOp := op.Record(gtx.Ops)
		clipStackOp := clip.Rect{Max: clipMax}.Push(gtx.Ops)
		callOp.Add(gtx.Ops)
		clipStackOp.Pop()
		callOp = macroOp.Stop()
		if clipMax.Y > 0 {
			newBaselineOffset := min(clipMax.Y, dims.Size.Y-dims.Baseline)
			dims.Baseline = clipMax.Y - newBaselineOffset
			if dims.Baseline < 0 {
				dims.Baseline = 0
			}
		}
		dims.Size = clipMax
	}
	return callOp, dims
}

func (c Container) layoutWidget(gtx layout.Context, cWidget layout.Widget) (layout.Dimensions, image.Point) {
	macroOp := op.Record(gtx.Ops)
	childGtx := gtx
	childGtx.Constraints.Min = image.Point{}
	childDimensions := cWidget(childGtx)
	callOp := macroOp.Stop()

	callOp, childDimensions = c.clipOverflow(gtx, callOp, childDimensions)
	childOffset := image.Point{
		X: c.getChildXOffset(gtx, childDimensions.Size.X),
		Y: c.getChildYOffset(gtx, childDimensions.Size.Y),
	}
	transformStack := op.Offset(childOffset).Push(gtx.Ops)
	callOp.Add(gtx.Ops)
	transformStack.Pop()

	availableHeight := getContainerHeight(gtx, c.FillDirection, childDimensions.Size.Y)
	childDimensions.Baseline = c.getChildBaseline(childDimensions, availableHeight)
	containerDimensions := c.getDimensions(gtx, childDimensions)

	return containerDimensions, childOffset
}

func (c Container) getChildXOffset(gtx layout.Context, childWidth int) int {
	containerWidth := getContainerWidth(gtx, c.FillDirection, childWidth)
	if c.Gravity == GravityTopStart || c.Gravity == GravityMiddleStart || c.Gravity == GravityBottomStart {
		if gtx.Locale.Direction == system.LTR {
			return 0
		} else {
			return containerWidth - childWidth
		}
	} else if c.Gravity == GravityTopCenter || c.Gravity == GravityMiddleCenter || c.Gravity == GravityBottomCenter {
		if gtx.Locale.Direction == system.LTR {
			return (containerWidth - childWidth) / 2
		} else {
			return (containerWidth - childWidth) / 2
		}
	} else if c.Gravity == GravityTopEnd || c.Gravity == GravityMiddleEnd || c.Gravity == GravityBottomEnd {
		if gtx.Locale.Direction == system.LTR {
			return containerWidth - childWidth
		} else {
			return 0
		}
	}
	panic("invalid gravity")
}

func (c Container) getChildYOffset(gtx layout.Context, childHeight int) int {
	containerHeight := getContainerHeight(gtx, c.FillDirection, childHeight)
	if c.Gravity == GravityTopStart || c.Gravity == GravityTopCenter || c.Gravity == GravityTopEnd {
		return 0
	} else if c.Gravity == GravityMiddleStart || c.Gravity == GravityMiddleCenter || c.Gravity == GravityMiddleEnd {
		return (containerHeight - childHeight) / 2
	} else if c.Gravity == GravityBottomStart || c.Gravity == GravityBottomCenter || c.Gravity == GravityBottomEnd {
		return containerHeight - childHeight
	}
	panic("invalid gravity")
}

func (c Container) getDimensions(gtx layout.Context, childDimensions layout.Dimensions) layout.Dimensions {
	switch c.FillDirection {
	case FillNone:
		// Nothing to do.
	case FillHorizontal:
		childDimensions.Size.X = gtx.Constraints.Max.X
	case FillVertical:
		childDimensions.Size.Y = gtx.Constraints.Max.Y
	case FillBoth:
		childDimensions.Size = gtx.Constraints.Max
	}
	if c.MinSize.X > 0 && childDimensions.Size.X < c.MinSize.X {
		childDimensions.Size.X = c.MinSize.X
	}
	if c.MinSize.Y > 0 && childDimensions.Size.Y < c.MinSize.Y {
		childDimensions.Size.Y = c.MinSize.Y
	}
	if childDimensions.Size.X < gtx.Constraints.Min.X {
		childDimensions.Size.X = gtx.Constraints.Min.X
	}
	if childDimensions.Size.Y < gtx.Constraints.Min.Y {
		childDimensions.Size.Y = gtx.Constraints.Min.Y
	}
	return layout.Dimensions{
		Size:     childDimensions.Size,
		Baseline: childDimensions.Baseline,
	}
}

func getContainerHeight(gtx layout.Context, fill FillDirection, childHeight int) int {
	if fill == FillVertical || fill == FillBoth {
		return gtx.Constraints.Max.Y
	}
	if childHeight > gtx.Constraints.Min.Y {
		return childHeight
	}
	return gtx.Constraints.Min.Y
}

func getContainerWidth(gtx layout.Context, fill FillDirection, childWidth int) int {
	if fill == FillHorizontal || fill == FillBoth {
		return gtx.Constraints.Max.X
	}
	if childWidth > gtx.Constraints.Min.X {
		return childWidth
	}
	return gtx.Constraints.Min.X
}
