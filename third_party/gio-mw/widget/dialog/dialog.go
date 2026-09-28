// SPDX-License-Identifier: Unlicense OR MIT

package dialog

import (
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"
	"gio-mw/widget/button"
	"gio-mw/widget/overlay"
	"image"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
)

const (
	iconSize            = unit.Dp(24)
	maximumWidth        = unit.Dp(560)
	maximumContentWidth = unit.Dp(560) - paddingContent*4
	minimumWidth        = unit.Dp(280)
	minimumContentWidth = unit.Dp(280) - paddingContent*4
	paddingContent      = unit.Dp(24)
	paddingIconHeadline = unit.Dp(16)
	paddingHeadlineBody = unit.Dp(16)
	paddingBodyActions  = unit.Dp(24)
	paddingButtons      = unit.Dp(8)
)

type kind int

const (
	kindBasic kind = iota
	kindFullScreen
)

type BasicStyle struct {
	Label          string
	Headline       string
	Subhead        string
	SupportingText string
	CancelText     string
	ConfirmText    string
	CancelButton   *button.Button
	ConfirmButton  *button.Button

	dTheme        *Theme
	wIcon         wdk.IconWidget
	overlayItemId int64
}

func (s *BasicStyle) AsOverlayItem() *overlay.Item {
	overlayItem := overlay.NewItem(s.Layout, block.GravityMiddleCenter).WithScrim()
	s.overlayItemId = overlayItem.GetId()
	return overlayItem
}

func (s *BasicStyle) GetLayoutItemId() int64 {
	return s.overlayItemId
}

func (s *BasicStyle) WithIcon(icon wdk.IconWidget) *BasicStyle {
	s.wIcon = icon
	return s
}

func (s *BasicStyle) Layout(gtx layout.Context) layout.Dimensions {
	// Draw contents and record macro and dimensions.
	macroOp := op.Record(gtx.Ops)
	contentDim := s.layoutContent(gtx)
	callOp := macroOp.Stop()

	// Build the dialog shape.
	baseBox := wdk.Box{
		Shape:    wdk.FromCornerShapesToken(gtx, s.dTheme.EnabledContainerShape),
		EndPoint: contentDim.Size,
	}

	// Draw the dialog shape.
	s.drawBackdrop(gtx, baseBox)

	finalLayerWidget := func(gtx layout.Context) layout.Dimensions {
		// Draw dialog contents.
		callOp.Add(gtx.Ops)
		return contentDim
	}
	return finalLayerWidget(gtx)
}

func (s *BasicStyle) drawBackdrop(gtx layout.Context, baseBoxShape wdk.Box) {
	// Draw the dialog elevation shadow.
	if s.dTheme.EnabledContainerElevation != token.ElevationLevel0 {
		dElevation := wdk.Elevation{
			Level:       s.dTheme.EnabledContainerElevation,
			ShadowColor: s.dTheme.EnabledContainerShadowColor,
		}
		dElevation.Layout(gtx, baseBoxShape)
	}
	paint.FillShape(
		gtx.Ops,
		s.dTheme.EnabledContainerColor.AsNRGBA(),
		baseBoxShape.Outline(gtx),
	)
}

func (s *BasicStyle) layoutContent(gtx layout.Context) layout.Dimensions {
	gtx.Constraints.Min.Y = 0
	if gtx.Constraints.Min.X < gtx.Dp(minimumContentWidth) {
		gtx.Constraints.Min.X = gtx.Dp(minimumContentWidth)
	}
	if gtx.Constraints.Max.X > gtx.Dp(maximumContentWidth) {
		gtx.Constraints.Max.X = gtx.Dp(maximumContentWidth)
	} else {
		gtx.Constraints.Max.X -= gtx.Dp(paddingContent * 2)
	}
	if gtx.Constraints.Max.X < gtx.Constraints.Min.X {
		gtx.Constraints.Max.X = gtx.Constraints.Min.X
	}

	return block.UniformPadding(paddingContent).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return layout.Stack{Alignment: layout.Center}.Layout(gtx, layout.Expanded(s.iconLayout))
			}),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				hAlignment := layout.W
				if s.wIcon != nil {
					hAlignment = layout.Center
				}
				return layout.Stack{Alignment: hAlignment}.Layout(gtx, layout.Expanded(s.headlineLayout))
			}),
			block.NewSegment(s.subheadLayout),
			block.NewSegment(s.labelLayout),
			block.NewSegment(func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return layout.Stack{Alignment: layout.E}.Layout(gtx, layout.Expanded(s.actionsLayout))
			}),
		)
	})
}

func (s *BasicStyle) iconLayout(gtx layout.Context) layout.Dimensions {
	if s.wIcon == nil {
		return layout.Dimensions{}
	}
	return block.Padding{
		Bottom: paddingIconHeadline,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min = image.Point{X: gtx.Dp(iconSize), Y: gtx.Dp(iconSize)}
		iconColor := s.dTheme.EnabledIconColor
		return s.wIcon(gtx, iconColor)
	})
}

func (s *BasicStyle) headlineLayout(gtx layout.Context) layout.Dimensions {
	headlineDim := layout.Dimensions{}
	headlineMacroOp := op.Record(gtx.Ops)
	if s.Headline != "" {
		headlineColor := s.dTheme.EnabledHeadlineColor
		tAlign := text.Start
		if s.wIcon != nil {
			tAlign = text.Middle
		}
		presentation := wdk.LabelStyle{
			Alignment: tAlign,
			Color:     headlineColor,
			Typestyle: token.TypestyleHeadlineSmall,
		}
		headlineDim = wdk.LayoutLabel(gtx, presentation, s.Headline)
	}
	headlineCallOp := headlineMacroOp.Stop()

	return block.Padding{
		Bottom: paddingHeadlineBody,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Render dialog headline.
		headlineCallOp.Add(gtx.Ops)

		return headlineDim
	})
}

func (s *BasicStyle) subheadLayout(gtx layout.Context) layout.Dimensions {
	if s.Subhead == "" {
		return layout.Dimensions{}
	}
	subheadColor := s.dTheme.EnabledSubheadColor
	presentation := wdk.LabelStyle{
		Alignment: text.Start,
		Color:     subheadColor,
		Typestyle: token.TypestyleHeadlineSmall,
	}
	return wdk.LayoutLabel(gtx, presentation, s.Subhead)
}

func (s *BasicStyle) labelLayout(gtx layout.Context) layout.Dimensions {
	if s.Label == "" {
		return layout.Dimensions{}
	}
	labelColor := s.dTheme.EnabledLabelColor
	presentation := wdk.LabelStyle{
		Alignment: text.Start,
		Color:     labelColor,
		Typestyle: token.TypestyleLabelMediumEmphasized,
	}
	return wdk.LayoutLabel(gtx, presentation, s.Label)
}

func (s *BasicStyle) actionsLayout(gtx layout.Context) layout.Dimensions {
	var widgets []block.Segment
	if s.CancelText != "" {
		cancelLayout := func(gtx layout.Context) layout.Dimensions {
			if s.CancelText == "" {
				return layout.Dimensions{}
			}
			return s.CancelButton.Layout(gtx, s.CancelText)
		}
		widgets = append(widgets, block.NewSegment(cancelLayout))
		widgets = append(widgets, block.NewHorizontalSpacer(paddingButtons))
	}
	confirmLayout := func(gtx layout.Context) layout.Dimensions {
		if s.ConfirmText == "" {
			panic("confirm button text is required")
		}
		return s.ConfirmButton.Layout(gtx, s.ConfirmText)
	}
	widgets = append(widgets, block.NewSegment(confirmLayout))

	return block.Padding{
		Top: paddingBodyActions,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisHorizontal,
			Overflow: block.OverflowClip,
		}.Layout(gtx, widgets...)
	})
}
