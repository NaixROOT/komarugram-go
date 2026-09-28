// SPDX-License-Identifier: Unlicense OR MIT

package input

import (
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"image"

	"gioui.org/f32"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
)

type widgetStyle struct {
	*Input
	shaper *text.Shaper
	theme  *Theme
}

func (s *widgetStyle) layout(gtx layout.Context) layout.Dimensions {
	// Layout input content.
	macroOp := op.Record(gtx.Ops)
	contentDims := s.layoutContent(gtx)
	callOp := macroOp.Stop()

	// TODO: Fix supportingTextHeight.
	bodySLineHeight := unit.Dp(12 * 1.3)
	supportingTextHeight := gtx.Dp(paddingSupportingTextTop) + gtx.Dp(bodySLineHeight)
	containerSize := image.Point{X: contentDims.Size.X, Y: contentDims.Size.Y - supportingTextHeight}
	baseBox := wdk.Box{
		Shape:    wdk.FromCornerShapesToken(gtx, s.theme.EnabledContainerShape),
		EndPoint: containerSize,
	}

	// Draw input shape.
	s.drawBackdrop(gtx, baseBox)

	// Draw input state layer.
	s.drawStateLayer(gtx, baseBox)

	// Draw the input content.
	callOp.Add(gtx.Ops)

	return contentDims
}

func (s *widgetStyle) drawBackdrop(gtx layout.Context, baseBox wdk.Box) {
	containerColor := s.theme.EnabledContainerColor
	activeIndicatorColor := s.theme.EnabledActiveIndicatorColor
	activeIndicatorHeight := float32(gtx.Dp(s.theme.EnabledActiveIndicatorHeight))
	switch s.Input.getWidgetState(gtx) {
	case Disabled:
		containerColor = s.theme.DisabledContainerColor.SetOpacity(s.theme.DisabledContainerOpacity)
		activeIndicatorColor = s.theme.DisabledActiveIndicatorColor.SetOpacity(s.theme.DisabledActiveIndicatorOpacity)
		activeIndicatorHeight = float32(gtx.Dp(s.theme.DisabledActiveIndicatorHeight))
	case Hovered:
		activeIndicatorHeight = float32(gtx.Dp(s.theme.HoveredActiveIndicatorHeight))
		activeIndicatorColor = s.theme.HoveredActiveIndicatorColor
	case Focused:
		activeIndicatorHeight = float32(gtx.Dp(s.theme.FocusedActiveIndicatorHeight))
		activeIndicatorColor = s.theme.FocusedActiveIndicatorColor
	case Error:
		activeIndicatorColor = s.theme.ErrorActiveIndicatorColor
	case ErrorHovered:
		activeIndicatorHeight = float32(gtx.Dp(s.theme.HoveredActiveIndicatorHeight))
		activeIndicatorColor = s.theme.ErrorHoveredActiveIndicatorColor
	case ErrorFocused:
		activeIndicatorHeight = float32(gtx.Dp(s.theme.FocusedActiveIndicatorHeight))
		activeIndicatorColor = s.theme.ErrorFocusedActiveIndicatorColor
	default:
		// Nothing to change!
	}

	anim := s.Input.getAnimation()
	containerColor = anim.container.Animate(gtx, containerColor)
	activeIndicatorColor = anim.indicator.Animate(gtx, activeIndicatorColor)
	activeIndicatorHeight = anim.indicatorHeight.Animate(gtx, activeIndicatorHeight)

	paint.FillShape(
		gtx.Ops,
		containerColor.AsNRGBA(),
		baseBox.Outline(gtx),
	)

	var p clip.Path
	p.Begin(gtx.Ops)
	activeIndicatorPosY := float32(baseBox.EndPoint.Y) - activeIndicatorHeight/2
	p.MoveTo(f32.Point{X: activeIndicatorHeight / 2, Y: activeIndicatorPosY})
	p.LineTo(f32.Point{X: float32(baseBox.EndPoint.X) - activeIndicatorHeight/2, Y: activeIndicatorPosY})
	activeIndicatorPathSpec := p.End()

	// Draw the active indicator line.
	paint.FillShape(
		gtx.Ops,
		activeIndicatorColor.AsNRGBA(),
		clip.Stroke{Path: activeIndicatorPathSpec, Width: activeIndicatorHeight}.Op(),
	)
}

// @see https://m3.material.io/foundations/interaction/states/state-layers
func (s *widgetStyle) drawStateLayer(gtx layout.Context, baseBox wdk.Box) {
	layerColor := s.theme.HoveredStateLayerColor.SetOpacity(0)
	switch s.Input.getWidgetState(gtx) {
	case Hovered:
		layerColor = s.theme.HoveredStateLayerColor.SetOpacity(s.theme.HoveredStateLayerOpacity)
	case ErrorHovered:
		layerColor = s.theme.ErrorHoveredStateLayerColor.SetOpacity(s.theme.ErrorHoveredStateLayerOpacity)
	default:
		// No state layer needed.
	}
	layerColor = s.Input.getAnimation().stateLayer.Animate(gtx, layerColor)
	if layerColor.A == 0 {
		return
	}
	paint.FillShape(
		gtx.Ops,
		layerColor.AsNRGBA(),
		baseBox.Outline(gtx),
	)
}

func (s *widgetStyle) layoutContent(gtx layout.Context) layout.Dimensions {
	leadingIconColor := s.theme.EnabledLeadingIconColor
	trailingIconColor := s.theme.EnabledTrailingIconColor
	supportingTextColor := s.theme.EnabledSupportingTextColor
	withText := s.Editor.GetText() != ""
	switch s.Input.getWidgetState(gtx) {
	case Disabled:
		leadingIconColor = s.theme.DisabledLeadingIconColor.SetOpacity(s.theme.DisabledLeadingIconOpacity)
		trailingIconColor = s.theme.DisabledTrailingIconColor.SetOpacity(s.theme.DisabledTrailingIconOpacity)
		supportingTextColor = s.theme.DisabledSupportingTextColor.SetOpacity(s.theme.DisabledSupportingTextOpacity)
	case Hovered:
		leadingIconColor = s.theme.HoveredLeadingIconColor
		trailingIconColor = s.theme.HoveredTrailingIconColor
		supportingTextColor = s.theme.HoveredSupportingTextColor
	case Focused:
		leadingIconColor = s.theme.FocusedLeadingIconColor
		trailingIconColor = s.theme.FocusedTrailingIconColor
		supportingTextColor = s.theme.FocusedSupportingTextColor
	case Error:
		leadingIconColor = s.theme.ErrorLeadingIconColor
		trailingIconColor = s.theme.ErrorTrailingIconColor
		supportingTextColor = s.theme.ErrorSupportingTextColor
	case ErrorHovered:
		if withText {
			leadingIconColor = s.theme.ErrorHoveredLeadingIconColor
			trailingIconColor = s.theme.ErrorHoveredTrailingIconColor
			supportingTextColor = s.theme.ErrorHoveredSupportingTextColor
		}
	case ErrorFocused:
		if withText {
			leadingIconColor = s.theme.ErrorFocusedLeadingIconColor
			trailingIconColor = s.theme.ErrorFocusedTrailingIconColor
			supportingTextColor = s.theme.ErrorFocusedSupportingTextColor
		}
	default:
		// Nothing to change.
	}

	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			leadingIconDims := layout.Dimensions{}
			var leadingIconCallOp op.CallOp
			if s.LeadingIcon != nil {
				macroOp := op.Record(gtx.Ops)
				leadingIconDims = block.Padding{Start: paddingLeadingIconStart}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					finalIconSize := gtx.Dp(s.theme.EnabledLeadingIconSize)
					gtx.Constraints = layout.Exact(image.Pt(finalIconSize, finalIconSize))
					return s.LeadingIcon(gtx, leadingIconColor)
				})
				leadingIconCallOp = macroOp.Stop()
			}
			trailingIconDims := layout.Dimensions{}
			var trailingIconCallOp op.CallOp
			if s.TrailingIcon != nil {
				macroOp := op.Record(gtx.Ops)
				trailingIconDims = block.Padding{End: paddingTrailingIconEnd}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					finalIconSize := gtx.Dp(s.theme.EnabledTrailingIconSize)
					gtx.Constraints = layout.Exact(image.Pt(finalIconSize, finalIconSize))
					return s.TrailingIcon(gtx, trailingIconColor)
				})
				trailingIconCallOp = macroOp.Stop()
			}
			return block.Line{
				Axis:     block.AxisHorizontal,
				Overflow: block.OverflowClip,
			}.Layout(gtx,
				block.NewSegment(func(gtx layout.Context) layout.Dimensions {
					leadingIconCallOp.Add(gtx.Ops)
					return leadingIconDims
				}).AlignMiddle(),
				block.NewFlexSegment(s.layoutMainContent).AlignMiddle(),
				block.NewSegment(func(gtx layout.Context) layout.Dimensions {
					trailingIconCallOp.Add(gtx.Ops)
					return trailingIconDims
				}).AlignMiddle(),
			)
		}),
		block.NewSegment(func(gtx layout.Context) layout.Dimensions {
			return block.Padding{
				Start: paddingStart,
				End:   paddingEnd,
				Top:   paddingSupportingTextTop,
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				tStyle := wdk.LabelStyle{
					Color:     supportingTextColor,
					Typestyle: token.TypestyleBodySmall,
				}
				return wdk.LayoutLabel(gtx, tStyle, s.SupportingText)
			})
		}),
	)
}

func (s *widgetStyle) layoutMainContent(gtx layout.Context) layout.Dimensions {
	inputTextColor := s.theme.EnabledInputTextColor
	labelTextColor := s.theme.EnabledLabelTextColor
	switch s.Input.getWidgetState(gtx) {
	case Disabled:
		inputTextColor = s.theme.DisabledInputTextColor.SetOpacity(s.theme.DisabledInputTextOpacity)
		labelTextColor = s.theme.DisabledLabelTextColor.SetOpacity(s.theme.DisabledLabelTextOpacity)
	case Hovered:
		inputTextColor = s.theme.HoveredInputTextColor
		labelTextColor = s.theme.HoveredLabelTextColor
	case Focused:
		inputTextColor = s.theme.FocusedInputTextColor
		labelTextColor = s.theme.FocusedLabelTextColor
	case Error:
		inputTextColor = s.theme.ErrorInputTextColor
		labelTextColor = s.theme.ErrorLabelTextColor
	case ErrorHovered:
		inputTextColor = s.theme.ErrorHoveredInputTextColor
		labelTextColor = s.theme.ErrorHoveredLabelTextColor
	case ErrorFocused:
		inputTextColor = s.theme.ErrorFocusedInputTextColor
		labelTextColor = s.theme.ErrorFocusedLabelTextColor
	default:
		// Nothing to change.
	}

	containerMinWidth := gtx.Dp(widthMin)
	if gtx.Constraints.Min.X > containerMinWidth {
		containerMinWidth = gtx.Constraints.Min.X
	}

	containerHeight := gtx.Dp(s.theme.EnabledContainerHeight)
	wdk.EnforceMin(&gtx, containerMinWidth, containerHeight)
	if s.Editor.SingleLine() {
		wdk.EnforceMax(&gtx, gtx.Constraints.Max.X, containerHeight)
	}
	return block.Container{
		Gravity: block.GravityMiddleStart,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Padding{
			Top:    paddingTop,
			Bottom: paddingBottom,
			Start:  paddingStart,
			End:    paddingEnd,
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			labelTextColor := s.Input.getAnimation().label.Animate(gtx, labelTextColor)
			return s.layoutMainContentForeground(gtx, inputTextColor, labelTextColor)
		})
	})
}

func (s *widgetStyle) layoutMainContentForeground(gtx layout.Context, inputTextColor token.MatColor, labelTextColor token.MatColor) layout.Dimensions {
	editorText := s.Editor.GetText()
	editorFocused := s.Editor.Focused(gtx)
	floatTarget := float32(1)
	if editorText == "" && !editorFocused {
		floatTarget = 0
	}
	floatProgress := s.Input.getAnimation().float.Animate(gtx, floatTarget)
	editorWidget := func(gtx layout.Context) layout.Dimensions {
		if s.Disabled {
			eStyle := wdk.LabelStyle{
				Color:     inputTextColor,
				Typestyle: token.TypestyleBodyLarge,
			}
			return wdk.LayoutLabel(gtx, eStyle, editorText)
		} else {
			materialTheme := wdk.GetMaterialTheme(gtx)
			typeInfo := materialTheme.Typescale[token.TypestyleBodyLarge]
			return s.Editor.Layout(gtx, wdk.EditorPresentation{
				Color:    inputTextColor,
				TypeInfo: typeInfo,
				Shaper:   s.shaper,
				// The caret appears once the label has moved out of its way.
				HideCaret: floatProgress != floatTarget,
			})
		}
	}
	if floatProgress != floatTarget {
		return s.layoutFloatingLabel(gtx, editorWidget, labelTextColor, floatTarget == 1, floatProgress)
	}
	if floatTarget == 0 {
		return block.Stack{
			Gravity: block.GravityMiddleStart,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.Y = 0
				tStyle := wdk.LabelStyle{
					Color:     labelTextColor,
					Typestyle: token.TypestyleBodyLarge,
				}
				return wdk.LayoutLabel(gtx, tStyle, s.LabelText)
			}),
			block.NewFlexSegment(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return editorWidget(gtx)
			}),
		)
	} else {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				tStyle := wdk.LabelStyle{
					Color:     labelTextColor,
					Typestyle: token.TypestyleBodySmall,
				}
				return wdk.LayoutLabel(gtx, tStyle, s.LabelText)
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return editorWidget(gtx)
			}),
		)
	}
}

// layoutFloatingLabel lays out the label while it moves between its resting
// place (large, centered) and its floating place (small, above the text).
// It reproduces both layouts of layoutMainContentForeground at the ends and
// draws the large label scaled in between.
func (s *widgetStyle) layoutFloatingLabel(gtx layout.Context, editorWidget layout.Widget, labelTextColor token.MatColor, floating bool, progress float32) layout.Dimensions {
	materialTheme := wdk.GetMaterialTheme(gtx)
	largeInfo := materialTheme.Typescale[token.TypestyleBodyLarge]
	smallInfo := materialTheme.Typescale[token.TypestyleBodySmall]
	area := gtx.Dp(s.theme.EnabledContainerHeight) - gtx.Dp(paddingTop) - gtx.Dp(paddingBottom)
	width := gtx.Constraints.Max.X

	labelGtx := gtx
	labelGtx.Constraints.Min = image.Point{}
	macroOp := op.Record(gtx.Ops)
	smallDims := wdk.LayoutLabel(labelGtx, wdk.LabelStyle{Typestyle: token.TypestyleBodySmall}, s.LabelText)
	macroOp.Stop()
	macroOp = op.Record(gtx.Ops)
	largeDims := wdk.LayoutLabel(labelGtx, wdk.LabelStyle{Color: labelTextColor, Typestyle: token.TypestyleBodyLarge}, s.LabelText)
	largeLabel := macroOp.Stop()

	editorGtx := gtx
	editorGtx.Constraints.Min.X = editorGtx.Constraints.Max.X
	editorGtx.Constraints.Min.Y = 0
	macroOp = op.Record(gtx.Ops)
	editorDims := editorWidget(editorGtx)
	editor := macroOp.Stop()

	// The editor sits where the target layout puts it.
	editorY := max((area-editorDims.Size.Y)/2, 0)
	if floating {
		editorY = smallDims.Size.Y
	}
	transformStack := op.Offset(image.Pt(0, editorY)).Push(gtx.Ops)
	editor.Add(gtx.Ops)
	transformStack.Pop()

	// The label scales from the large size to the small one, keeping the
	// centers of the resting and floating lines aligned.
	scale := 1 + (float32(smallInfo.Size)/float32(largeInfo.Size)-1)*progress
	restingY := float32(area-largeDims.Size.Y) / 2
	floatingY := float32(smallDims.Size.Y)/2 - float32(largeDims.Size.Y)*scale/2
	y := restingY + (floatingY-restingY)*progress
	x := float32(0)
	if gtx.Locale.Direction == system.RTL {
		x = float32(width) - float32(largeDims.Size.X)*scale
	}
	transformStack = op.Affine(f32.AffineId().Scale(f32.Point{}, f32.Pt(scale, scale)).Offset(f32.Pt(x, y))).Push(gtx.Ops)
	largeLabel.Add(gtx.Ops)
	transformStack.Pop()

	return layout.Dimensions{Size: image.Pt(width, max(area, editorY+editorDims.Size.Y))}
}
