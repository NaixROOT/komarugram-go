// SPDX-License-Identifier: Unlicense OR MIT

package badge

import (
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"image"
	"strconv"

	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
)

type Badge struct {
	Visible      bool
	MaxMagnitude uint8
}

func (b *Badge) OnIconWidget(widget wdk.IconWidget, itemCount uint) wdk.IconWidget {
	if !b.Visible {
		return widget
	}

	return func(gtx layout.Context, foreground token.MatColor) layout.Dimensions {
		macroOp := op.Record(gtx.Ops)
		iconDimensions := widget(gtx, foreground)
		callOp := macroOp.Stop()

		widgetTheme := BuildTheme(gtx)
		callOp.Add(gtx.Ops)
		if itemCount == 0 {
			b.drawSmallBadge(gtx, widgetTheme, iconDimensions)
		} else {
			infoMacroOp := op.Record(gtx.Ops)
			infoDimensions := b.layoutInfoBadge(gtx, itemCount, widgetTheme)
			infoCallOp := infoMacroOp.Stop()

			// TODO: Fix overflow clipping.
			xOffset := 0
			if gtx.Locale.Direction == system.RTL {
				xOffset = iconDimensions.Size.X/2 - infoDimensions.Size.X
			} else {
				xOffset = iconDimensions.Size.X / 2
			}
			yOffset := iconDimensions.Size.Y/2 - infoDimensions.Size.Y
			offsetPoint := image.Pt(xOffset, yOffset)
			transformStack := op.Offset(offsetPoint).Push(gtx.Ops)
			infoCallOp.Add(gtx.Ops)
			transformStack.Pop()
		}
		return iconDimensions
	}
}

func (b *Badge) drawSmallBadge(gtx layout.Context, widgetTheme *Theme, iconDimensions layout.Dimensions) {
	badgeSize := gtx.Dp(widgetTheme.BadgeSize)
	xOffset := 0
	if gtx.Locale.Direction == system.RTL {
		xOffset = 0
	} else {
		xOffset = iconDimensions.Size.X - badgeSize
	}
	startPoint := image.Point{X: xOffset, Y: 0}
	endPoint := startPoint.Add(image.Pt(badgeSize, badgeSize))
	boxShape := wdk.Box{
		Shape: wdk.UniformCornerShapes(wdk.CornerShape{
			Kind:        wdk.CornerKindRound,
			AdaptToSize: true,
		}),
		StartPoint: startPoint,
		EndPoint:   endPoint,
	}
	paint.FillShape(gtx.Ops, widgetTheme.BadgeColor.AsNRGBA(), boxShape.Outline(gtx))
}

func (b *Badge) layoutInfoBadge(gtx layout.Context, itemCount uint, widgetTheme *Theme) layout.Dimensions {
	return block.Background{
		CornerShapes: wdk.FromCornerShapesToken(gtx, widgetTheme.BadgeLargeShape),
		Color:        widgetTheme.BadgeLargeColor,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Padding{
			Start: unit.Dp(4),
			End:   unit.Dp(4),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Dp(8)
			infoText := ""
			switch b.MaxMagnitude {
			case 1:
				if itemCount > 9 {
					infoText = "9+"
				} else {
					infoText = strconv.Itoa(int(itemCount))
				}
			case 2:
				if itemCount > 99 {
					infoText = "99+"
				} else {
					infoText = strconv.Itoa(int(itemCount))
				}
			default:
				if itemCount > 999 {
					infoText = "999+"
				} else {
					infoText = strconv.Itoa(int(itemCount))
				}
			}
			presentation := wdk.LabelStyle{
				Alignment: text.Middle,
				Color:     widgetTheme.BadgeLargeLabelTextColor,
				MaxLines:  1,
				Typestyle: widgetTheme.BadgeLargeLabelTypestyle,
			}
			return wdk.LayoutLabel(gtx, presentation, infoText)
		})
	})
}
