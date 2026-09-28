// SPDX-License-Identifier: Unlicense OR MIT

package search

import (
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"image"

	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
)

type widgetStyle struct {
	*Search
	shaper *text.Shaper
	theme  *Theme
}

func (s *widgetStyle) layout(gtx layout.Context) layout.Dimensions {
	containerWidthMin := gtx.Dp(widthMin - paddingHorizontal*2)
	containerWidthMax := gtx.Dp(widthMax - paddingHorizontal*2)
	if containerWidthMax > gtx.Constraints.Max.X {
		containerWidthMax = gtx.Constraints.Max.X
	}
	wdk.EnforceWidth(&gtx, containerWidthMin, containerWidthMax)

	// Layout search content.
	macroOp := op.Record(gtx.Ops)
	contentDims := s.layoutContent(gtx)
	callOp := macroOp.Stop()

	// Build the input shape.
	baseBox := wdk.Box{
		Shape:    wdk.FromCornerShapesToken(gtx, s.theme.EnabledContainerShape),
		EndPoint: contentDims.Size,
	}

	// Draw the input shape.
	s.drawBackdrop(gtx, baseBox)

	// Draw the input state layer.
	s.drawStateLayer(gtx, baseBox)

	// Draw the search content.
	callOp.Add(gtx.Ops)

	return contentDims
}

func (s *widgetStyle) drawBackdrop(gtx layout.Context, baseBox wdk.Box) {
	containerColor := s.theme.EnabledContainerColor
	paint.FillShape(
		gtx.Ops,
		containerColor.AsNRGBA(),
		baseBox.Outline(gtx),
	)
}

func (s *widgetStyle) drawStateLayer(gtx layout.Context, box wdk.Box) {
}

func (s *widgetStyle) layoutContent(gtx layout.Context) layout.Dimensions {
	return block.Line{
		Axis:     block.AxisHorizontal,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			if !s.hasLeadingIcon() {
				panic("LeadingIcon is required for a search widget.")
			}
			return s.layoutLeadingIcon(gtx)
		}).AlignMiddle(),
		block.NewFlexSegment(func(gtx layout.Context) layout.Dimensions {
			//return debug.Layout(gtx, s.shaper, s.layoutMainContent)
			return s.layoutMainContent(gtx)
		}).AlignMiddle(),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			if s.hasTrailingButton() {
				return s.layoutTrailingIcon(gtx)
			}
			return layout.Dimensions{}
		}).AlignMiddle(),
	)
}

func (s *widgetStyle) layoutLeadingIcon(gtx layout.Context) layout.Dimensions {
	// TODO: Change search icon color on hover and focus
	leadingIconColor := s.theme.EnabledLeadingIconColor
	leadingIcon := func(gtx layout.Context) layout.Dimensions {
		p := block.Padding{Start: paddingHorizontal}
		return p.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			finalIconSize := gtx.Dp(iconSize)
			gtx.Constraints = layout.Exact(image.Pt(finalIconSize, finalIconSize))
			return s.LeadingIcon.Icon(gtx, leadingIconColor)
		})
	}
	if s.hasLeadingButton() {
		semantic.Button.Add(gtx.Ops)
		semantic.LabelOp(s.LeadingIcon.Label).Add(gtx.Ops)
		if s.Search.LeadingIcon.Clickable.Hovered() {
			pointer.CursorPointer.Add(gtx.Ops)
		}
		return s.Search.LeadingIcon.Clickable.Layout(gtx, leadingIcon)
	} else {
		return leadingIcon(gtx)
	}
}

func (s *widgetStyle) hasLeadingIcon() bool {
	return s.LeadingIcon.Icon != nil
}

func (s *widgetStyle) hasLeadingButton() bool {
	return s.LeadingIcon.Icon != nil && s.LeadingIcon.Label != ""
}

func (s *widgetStyle) hasTrailingButton() bool {
	return s.TrailingIcon.Icon != nil && s.TrailingIcon.Label != ""
}

func (s *widgetStyle) layoutTrailingIcon(gtx layout.Context) layout.Dimensions {
	if !s.hasTrailingButton() {
		return layout.Dimensions{}
	}
	// TODO: Change search icon color on hover and focus
	trailingIconColor := s.theme.EnabledTrailingIconColor
	if s.Search.TrailingIcon.Clickable.Hovered() {
		pointer.CursorPointer.Add(gtx.Ops)
	}
	semantic.Button.Add(gtx.Ops)
	semantic.LabelOp(s.TrailingIcon.Label).Add(gtx.Ops)
	return s.Search.TrailingIcon.Clickable.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		p := block.Padding{End: paddingHorizontal}
		return p.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			finalIconSize := gtx.Dp(iconSize)
			gtx.Constraints = layout.Exact(image.Pt(finalIconSize, finalIconSize))
			return s.Search.TrailingIcon.Icon(gtx, trailingIconColor)
		})
	})
}

func (s *widgetStyle) layoutMainContent(gtx layout.Context) layout.Dimensions {
	supportingTextColor := s.theme.EnabledSupportingTextColor
	inputTextColor := s.theme.EnabledInputTextColor
	switch s.Search.getWidgetState(gtx) {
	case Hovered:
		supportingTextColor = s.theme.HoveredSupportingTextColor
	case Pressed:
		supportingTextColor = s.theme.PressedSupportingTextColor
	default:
		// Nothing to change.
	}

	containerHeight := gtx.Dp(s.theme.EnabledContainerHeight)
	wdk.EnforceHeight(&gtx, containerHeight, containerHeight)

	return block.Padding{
		Start: paddingHorizontal,
		End:   paddingHorizontal,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		editorWidget := func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			materialTheme := wdk.GetMaterialTheme(gtx)
			typeInfo := materialTheme.Typescale[token.TypestyleBodyLarge]
			return s.editor.Layout(gtx, wdk.EditorPresentation{
				Color:    inputTextColor,
				TypeInfo: typeInfo,
				Shaper:   s.shaper,
			})
		}
		// The hint stays until something is typed, focused or not.
		if s.editor.GetText() == "" {
			return block.Stack{
				Gravity: block.GravityMiddleStart,
			}.Layout(gtx,
				block.NewSegment(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.Y = 0
					presentation := wdk.LabelStyle{
						Color:     supportingTextColor,
						Typestyle: token.TypestyleBodyLarge,
					}
					return wdk.LayoutLabel(gtx, presentation, s.SupportingText)
				}),
				block.NewFlexSegment(func(gtx layout.Context) layout.Dimensions {
					// Centered like the hint, so that the caret sits on its line.
					return block.Container{
						Gravity: block.GravityMiddleStart,
					}.Layout(gtx, editorWidget)
				}),
			)
		}

		return block.Container{
			Gravity: block.GravityMiddleStart,
		}.Layout(gtx, editorWidget)
	})
}
