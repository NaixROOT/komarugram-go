// SPDX-License-Identifier: Unlicense OR MIT

package components

import (
	"gio-mw/exp/examples"
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type ColorBox struct {
	Name       string
	ColorSet   *token.MatColorSet
	JustFirst  bool
	JustSecond bool
	SkipPrefix bool
}

func NewColorBox(name string, colorSet *token.MatColorSet) ColorBox {
	return ColorBox{
		Name:     name,
		ColorSet: colorSet,
	}
}

func NewColorBoxFromColors(name string, color1, color2 *token.MatColor) ColorBox {
	return ColorBox{
		Name: name,
		ColorSet: &token.MatColorSet{
			Color:   *color1,
			OnColor: *color2,
		},
	}
}

func (cb *ColorBox) Layout(gtx layout.Context) layout.Dimensions {
	if cb.JustFirst {
		return cb.FirstBoxLayout(gtx)
	}
	if cb.JustSecond {
		return cb.SecondBoxLayout(gtx)
	}
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(cb.FirstBoxLayout),
		block.NewSegment(cb.SecondBoxLayout),
	)
}

func (cb *ColorBox) FirstBoxLayout(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.Y = 56
	return Box(gtx, func(gtx layout.Context) layout.Dimensions {
		lStyle := wdk.LabelStyle{
			Color:     cb.ColorSet.OnColor,
			MaxLines:  3,
			Typestyle: token.TypestyleLabelLargeEmphasized,
		}
		return wdk.LayoutLabel(gtx, lStyle, cb.Name)
	}, cb.ColorSet.Color)
}

func (cb *ColorBox) SecondBoxLayout(gtx layout.Context) layout.Dimensions {
	txt := cb.Name
	if !cb.SkipPrefix {
		txt = "On " + cb.Name
	}
	gtx.Constraints.Min.Y = 40
	return Box(gtx, func(gtx layout.Context) layout.Dimensions {
		lStyle := wdk.LabelStyle{
			Color:     cb.ColorSet.Color,
			MaxLines:  3,
			Typestyle: token.TypestyleLabelLargeEmphasized,
		}
		return wdk.LayoutLabel(gtx, lStyle, txt)
	}, cb.ColorSet.OnColor)
}

type DualColorBox struct {
	Name1     string
	ColorSet1 *token.MatColorSet
	Name2     string
	ColorSet2 *token.MatColorSet
}

func (cb *DualColorBox) Layout(gtx layout.Context) layout.Dimensions {
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(cb.FirstBoxLayout),
		block.NewSegment(cb.SecondBoxLayout),
	)
}

func (cb *DualColorBox) FirstBoxLayout(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.Y = 56
	maxWidth := gtx.Constraints.Max.X

	var widgets []block.Segment
	widgets = append(widgets,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = maxWidth / 2
			return Box(gtx, func(gtx layout.Context) layout.Dimensions {
				lStyle := wdk.LabelStyle{
					Color:     cb.ColorSet1.OnColor,
					MaxLines:  3,
					Typestyle: token.TypestyleLabelLargeEmphasized,
				}
				return wdk.LayoutLabel(gtx, lStyle, cb.Name1)
			}, cb.ColorSet1.Color)
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = maxWidth / 2
			return Box(gtx, func(gtx layout.Context) layout.Dimensions {
				lStyle := wdk.LabelStyle{
					Color:     cb.ColorSet2.OnColor,
					MaxLines:  3,
					Typestyle: token.TypestyleLabelLargeEmphasized,
				}
				return wdk.LayoutLabel(gtx, lStyle, cb.Name2)
			}, cb.ColorSet2.Color)
		}),
	)

	return block.Line{
		Axis:     block.AxisHorizontal,
		Overflow: block.OverflowClip,
	}.Layout(gtx, widgets...)
}

func (cb *DualColorBox) SecondBoxLayout(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.Y = 40

	var widgets []block.Segment
	widgets = append(widgets,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			txt := "On " + cb.Name1
			return Box(gtx, func(gtx layout.Context) layout.Dimensions {
				lStyle := wdk.LabelStyle{
					Color:     cb.ColorSet1.Color,
					MaxLines:  3,
					Typestyle: token.TypestyleLabelLargeEmphasized,
				}
				return wdk.LayoutLabel(gtx, lStyle, txt)
			}, cb.ColorSet1.OnColor)
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			txt := "On " + cb.Name2
			return Box(gtx, func(gtx layout.Context) layout.Dimensions {
				lStyle := wdk.LabelStyle{
					Color:     cb.ColorSet2.Color,
					MaxLines:  3,
					Typestyle: token.TypestyleLabelLargeEmphasized,
				}
				return wdk.LayoutLabel(gtx, lStyle, txt)
			}, cb.ColorSet2.OnColor)
		}),
	)
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx, widgets...)
}

type ColorBoxGroup struct {
	ColorBox     *ColorBox
	ContainerBox *ColorBox
}

func NewColorBoxGroup(name string, color, container *token.MatColorSet) ColorBoxGroup {
	colorBox := NewColorBox(name, color)
	containerBox := NewColorBox(name+" Container", container)
	return ColorBoxGroup{
		ColorBox:     &colorBox,
		ContainerBox: &containerBox,
	}
}

func (bg *ColorBoxGroup) Layout(gtx layout.Context) layout.Dimensions {
	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(bg.ColorBox.Layout),
		block.NewVerticalSpacer(examples.SpacingTiny),
		block.NewSegment(bg.ContainerBox.Layout),
	)
}

func Box(gtx layout.Context, cWidget layout.Widget, cColor token.MatColor) layout.Dimensions {
	// Layout the box content and record the size.
	macroOp := op.Record(gtx.Ops)
	contentDim := block.UniformPadding(examples.SpacingSmall).Layout(gtx, cWidget)
	callOp := macroOp.Stop()

	// Increase the height of the box to fill the requested height.
	if contentDim.Size.Y < gtx.Constraints.Min.Y {
		contentDim.Size.Y = gtx.Constraints.Min.Y
	}

	// Increase the width of the box to fill the available width.
	if contentDim.Size.X < gtx.Constraints.Max.X {
		contentDim.Size.X = gtx.Constraints.Max.X
	}

	// Draw the background.
	paint.FillShape(
		gtx.Ops,
		cColor.AsNRGBA(),
		clip.Rect{Max: contentDim.Size}.Op(),
	)
	// Draw the content and return the size.
	callOp.Add(gtx.Ops)
	return contentDim
}
