// SPDX-License-Identifier: Unlicense OR MIT

package components

import (
	"fmt"
	"gio-mw/exp/examples"
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"image"
	"strings"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
)

type LabelBox struct {
	styleInfo token.TypeInfo
	typestyle token.Typestyle
}

func NewLabelBox(typestyle token.Typestyle) block.Segment {
	return block.Segment{
		BaseSize: unit.Dp(256),
		Flex:     1,
		Widget: func(gtx layout.Context) layout.Dimensions {
			materialTheme := wdk.GetMaterialTheme(gtx)
			b := LabelBox{
				styleInfo: materialTheme.Typescale[typestyle],
				typestyle: typestyle,
			}
			return block.UniformPadding(examples.SpacingTiny).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return widget.Border{
					Color:        materialTheme.Scheme.Outline.AsNRGBA(),
					CornerRadius: unit.Dp(4),
					Width:        unit.Dp(1),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return block.UniformPadding(examples.SpacingSmall).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return block.Line{
							Axis:     block.AxisVertical,
							Overflow: block.OverflowClip,
						}.Layout(gtx,
							block.NewSegment(getSizeBox(b)),
							block.NewSegment(getStyleBox(b)),
						)
					})
				})
			})
		},
	}
}

func getSizeBox(b LabelBox) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		lStyle := wdk.LabelStyle{
			Typestyle: token.TypestyleLabelMedium,
		}
		txt := fmt.Sprintf("%v", b.styleInfo.Size)
		return wdk.LayoutLabel(gtx, lStyle, txt)
	}
}

func getStyleBox(b LabelBox) layout.Widget {
	nameParts := strings.Split(b.styleInfo.Name, " ")
	groupW := getGroupBox(nameParts, b)
	variantW := getVariantBox(nameParts, b)
	return func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisHorizontal,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(groupW),
			block.NewSegment(variantW),
		)
	}
}

func getGroupBox(nameParts []string, b LabelBox) layout.Widget {
	styleGroup := nameParts[0]
	return func(gtx layout.Context) layout.Dimensions {
		return block.Padding{End: examples.SpacingSmall}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				materialTheme := wdk.GetMaterialTheme(gtx)
				presentation := wdk.LabelStyle{
					Typestyle: b.typestyle,
					Color:     materialTheme.Scheme.Secondary.Color,
				}
				return wdk.LayoutLabel(gtx, presentation, styleGroup)
			},
		)
	}
}

func getVariantBox(nameParts []string, b LabelBox) layout.Widget {
	return func(gtx layout.Context) layout.Dimensions {
		materialTheme := wdk.GetMaterialTheme(gtx)
		colorSet := materialTheme.Scheme.TertiaryContainer
		textW := func(gtx layout.Context) layout.Dimensions {
			var styleVariant string
			switch nameParts[1] {
			case "Large":
				styleVariant = "L"
			case "Medium":
				styleVariant = "M"
			case "Small":
				styleVariant = "S"
			}
			inset := block.UniformPadding(examples.SpacingTiny)
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lStyle := wdk.LabelStyle{
					Color:     colorSet.OnColor,
					Alignment: text.End,
					MaxLines:  1,
					Typestyle: token.TypestyleLabelMedium,
				}
				return wdk.LayoutLabel(gtx, lStyle, styleVariant)
			})
		}
		fillW := func(gtx layout.Context) layout.Dimensions {
			rr := gtx.Dp(2)
			defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, rr).Push(gtx.Ops).Pop()
			paint.Fill(gtx.Ops, colorSet.Color.AsNRGBA())
			return layout.Dimensions{Size: gtx.Constraints.Min}
		}
		return layout.Stack{Alignment: layout.Center}.Layout(gtx, layout.Expanded(fillW), layout.Stacked(textW))
	}
}
