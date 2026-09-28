// SPDX-License-Identifier: Unlicense OR MIT

package colors

import (
	"gio-mw/examples/kitchen/components"
	"gio-mw/exp/examples"
	"gio-mw/exp/router"
	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/wdk/block"

	"gioui.org/layout"
)

type Page struct {
	// TODO: Remove this reference.
	Theme *token.Theme
}

func NewPage() router.PageWidget {
	return &Page{}
}

func (p *Page) IsWide() bool {
	return false
}

func (p *Page) Update(gtx layout.Context) {
	// Nothing to do.
}

func (p *Page) View(gtx layout.Context) layout.Dimensions {
	p.Theme = wdk.GetMaterialTheme(gtx)
	return block.UniformPadding(examples.SpacingMedium).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisVertical,
			Overflow: block.OverflowClip,
		}.Layout(gtx,
			block.NewSegment(p.mainColorsRowLayout),
			block.NewSegment(p.fixedColorsRowLayout),
			block.NewSegment(p.backdropColorsRowLayout),
			block.NewSegment(p.surfaceColorsRowLayout),
			block.NewSegment(p.surfaceContainerColorsRowLayout),
			block.NewSegment(p.onSurfaceColorsRowLayout),
		)
	})
}

func (p *Page) mainColorsRowLayout(gtx layout.Context) layout.Dimensions {
	primaryGroup := components.NewColorBoxGroup("Primary", &p.Theme.Scheme.Primary, &p.Theme.Scheme.PrimaryContainer)
	secondaryGroup := components.NewColorBoxGroup("Secondary", &p.Theme.Scheme.Secondary, &p.Theme.Scheme.SecondaryContainer)
	tertiaryGroup := components.NewColorBoxGroup("Tertiary", &p.Theme.Scheme.Tertiary, &p.Theme.Scheme.TertiaryContainer)
	errorGroup := components.NewColorBoxGroup("Error", &p.Theme.Scheme.Error, &p.Theme.Scheme.ErrorContainer)

	var widgets []block.Segment
	widgets = append(widgets,
		block.Segment{Flex: 1, BaseSize: 192, Widget: primaryGroup.Layout},
		block.NewHorizontalSpacer(examples.SpacingTiny),
		block.Segment{Flex: 1, BaseSize: 192, Widget: secondaryGroup.Layout},
		block.NewHorizontalSpacer(examples.SpacingTiny),
		block.Segment{Flex: 1, BaseSize: 192, Widget: tertiaryGroup.Layout},
		block.NewHorizontalSpacer(examples.SpacingMedium),
		block.Segment{Flex: 1, BaseSize: 192, Widget: errorGroup.Layout},
	)

	return block.UniformPadding(examples.SpacingSmall).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisHorizontal,
			Overflow: block.OverflowWrap,
			Expand:   true,
		}.Layout(gtx, widgets...)
	})
}

func (p *Page) fixedColorsRowLayout(gtx layout.Context) layout.Dimensions {
	primaryFixedBox := components.DualColorBox{
		Name1:     "Primary Fixed",
		ColorSet1: &p.Theme.Scheme.PrimaryFixed,
		Name2:     "Primary Fixed Variant",
		ColorSet2: &p.Theme.Scheme.PrimaryFixedVariant,
	}
	secondaryFixedBox := components.DualColorBox{
		Name1:     "Secondary Fixed",
		ColorSet1: &p.Theme.Scheme.SecondaryFixed,
		Name2:     "Secondary Fixed Variant",
		ColorSet2: &p.Theme.Scheme.SecondaryFixedVariant,
	}
	tertiaryFixedBox := components.DualColorBox{
		Name1:     "Tertiary Fixed",
		ColorSet1: &p.Theme.Scheme.TertiaryFixed,
		Name2:     "Tertiary Fixed Variant",
		ColorSet2: &p.Theme.Scheme.TertiaryFixedVariant,
	}
	emptyWidget := func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{}
	}

	var widgets []block.Segment
	widgets = append(widgets,
		block.Segment{Flex: 1, BaseSize: 192, Widget: primaryFixedBox.Layout},
		block.NewHorizontalSpacer(examples.SpacingTiny),
		block.Segment{Flex: 1, BaseSize: 192, Widget: secondaryFixedBox.Layout},
		block.NewHorizontalSpacer(examples.SpacingTiny),
		block.Segment{Flex: 1, BaseSize: 192, Widget: tertiaryFixedBox.Layout},
		block.NewHorizontalSpacer(examples.SpacingMedium),
		block.Segment{Flex: 1, BaseSize: 192, Widget: emptyWidget},
	)
	return block.UniformPadding(examples.SpacingSmall).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisHorizontal,
			Overflow: block.OverflowWrap,
			Expand:   true,
		}.Layout(gtx, widgets...)
	})
}

func (p *Page) backdropColorsRowLayout(gtx layout.Context) layout.Dimensions {
	backgroundBox := components.NewColorBox("Background", &p.Theme.Scheme.Background)
	backgroundBox.JustFirst = true
	scrimBoxColor := token.MatColorSet{
		Color: p.Theme.Scheme.Scrim,
		OnColor: token.MatColor{
			R: 0xFF - p.Theme.Scheme.Scrim.R,
			G: 0xFF - p.Theme.Scheme.Scrim.G,
			B: 0xFF - p.Theme.Scheme.Scrim.B,
			A: 0xFF,
		},
	}
	scrimBox := components.NewColorBox("Scrim", &scrimBoxColor)
	scrimBox.JustFirst = true
	shadowBoxColor := token.MatColorSet{
		Color: p.Theme.Scheme.Shadow,
		OnColor: token.MatColor{
			R: 0xFF - p.Theme.Scheme.Shadow.R,
			G: 0xFF - p.Theme.Scheme.Shadow.G,
			B: 0xFF - p.Theme.Scheme.Shadow.B,
			A: 0xFF,
		},
	}
	shadowBox := components.NewColorBox("Shadow", &shadowBoxColor)
	shadowBox.JustFirst = true

	var widgets []block.Segment
	widgets = append(widgets,
		block.Segment{Flex: 1, BaseSize: 128, Widget: backgroundBox.Layout},
		block.NewHorizontalSpacer(examples.SpacingTiny),
		block.Segment{Flex: 1, BaseSize: 128, Widget: scrimBox.Layout},
		block.NewHorizontalSpacer(examples.SpacingTiny),
		block.Segment{Flex: 1, BaseSize: 128, Widget: shadowBox.Layout},
	)

	return block.UniformPadding(examples.SpacingSmall).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisHorizontal,
			Overflow: block.OverflowWrap,
			Expand:   true,
		}.Layout(gtx, widgets...)
	})
}

func (p *Page) surfaceColorsRowLayout(gtx layout.Context) layout.Dimensions {
	dimBox := components.NewColorBoxFromColors("Surface Dim", &p.Theme.Scheme.SurfaceDim, &p.Theme.Scheme.Surface.OnColor)
	dimBox.JustFirst = true
	normalBox := components.NewColorBox("Surface", &p.Theme.Scheme.Surface)
	normalBox.JustFirst = true
	brightBox := components.NewColorBoxFromColors("Surface Bright", &p.Theme.Scheme.SurfaceBright, &p.Theme.Scheme.Surface.OnColor)
	brightBox.JustFirst = true

	var widgets []block.Segment
	widgets = append(widgets,
		block.Segment{Flex: 1, BaseSize: 128, Widget: dimBox.Layout},
		block.NewHorizontalSpacer(examples.SpacingTiny),
		block.Segment{Flex: 1, BaseSize: 128, Widget: normalBox.Layout},
		block.NewHorizontalSpacer(examples.SpacingTiny),
		block.Segment{Flex: 1, BaseSize: 128, Widget: brightBox.Layout},
		block.NewHorizontalSpacer(examples.SpacingMedium),
		block.Segment{Flex: 1, BaseSize: 128, Widget: p.inverseSurfaceBoxLayout},
	)

	return block.UniformPadding(examples.SpacingSmall).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisHorizontal,
			Overflow: block.OverflowWrap,
			Expand:   true,
		}.Layout(gtx, widgets...)
	})
}

func (p *Page) inverseSurfaceBoxLayout(gtx layout.Context) layout.Dimensions {
	inverseBox := components.NewColorBox("Inverse Surface", &p.Theme.Scheme.InverseSurface)
	inversePrimaryBox := components.NewColorBoxFromColors("Inverse Primary", &p.Theme.Scheme.InversePrimary, &p.Theme.Scheme.Surface.OnColor)
	inversePrimaryBox.JustFirst = true

	return block.Line{
		Axis:     block.AxisVertical,
		Overflow: block.OverflowClip,
	}.Layout(gtx,
		block.NewSegment(inverseBox.Layout),
		block.NewSegment(examples.Spacer(examples.SpacingTiny)),
		block.NewSegment(inversePrimaryBox.Layout),
	)
}

func (p *Page) surfaceContainerColorsRowLayout(gtx layout.Context) layout.Dimensions {
	lowestBox := components.NewColorBoxFromColors("Surface Container Lowest", &p.Theme.Scheme.SurfaceContainerLowest, &p.Theme.Scheme.Surface.OnColor)
	lowestBox.JustFirst = true
	lowBox := components.NewColorBoxFromColors("Surface Container Low", &p.Theme.Scheme.SurfaceContainerLow, &p.Theme.Scheme.Surface.OnColor)
	lowBox.JustFirst = true
	normalBox := components.NewColorBoxFromColors("Surface Container", &p.Theme.Scheme.SurfaceContainer, &p.Theme.Scheme.Surface.OnColor)
	normalBox.JustFirst = true
	highBox := components.NewColorBoxFromColors("Surface Container High", &p.Theme.Scheme.SurfaceContainerHigh, &p.Theme.Scheme.Surface.OnColor)
	highBox.JustFirst = true
	highestBox := components.NewColorBoxFromColors("Surface Container Highest", &p.Theme.Scheme.SurfaceContainerHighest, &p.Theme.Scheme.Surface.OnColor)
	highestBox.JustFirst = true

	var widgets []block.Segment
	widgets = append(widgets,
		block.Segment{Flex: 1, BaseSize: 128, Widget: lowestBox.Layout},
		block.Segment{Flex: 1, BaseSize: 128, Widget: lowBox.Layout},
		block.Segment{Flex: 1, BaseSize: 128, Widget: normalBox.Layout},
		block.Segment{Flex: 1, BaseSize: 128, Widget: highBox.Layout},
		block.Segment{Flex: 1, BaseSize: 128, Widget: highestBox.Layout},
	)

	return block.UniformPadding(examples.SpacingSmall).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisHorizontal,
			Overflow: block.OverflowWrap,
			Expand:   true,
		}.Layout(gtx, widgets...)
	})
}

func (p *Page) onSurfaceColorsRowLayout(gtx layout.Context) layout.Dimensions {
	surfaceBox := components.NewColorBox("Surface", &p.Theme.Scheme.Surface)
	surfaceBox.JustSecond = true
	surfaceVariantBox := components.NewColorBox("Surface Variant", &p.Theme.Scheme.SurfaceVariant)
	surfaceVariantBox.JustSecond = true
	outlineBox := components.NewColorBoxFromColors("Outline", &p.Theme.Scheme.Background.Color, &p.Theme.Scheme.Outline)
	outlineBox.JustSecond = true
	outlineBox.SkipPrefix = true
	outlineVariantBox := components.NewColorBoxFromColors("Outline Variant", &p.Theme.Scheme.Background.Color, &p.Theme.Scheme.OutlineVariant)
	outlineVariantBox.JustSecond = true
	outlineVariantBox.SkipPrefix = true

	var widgets []block.Segment
	widgets = append(widgets,
		block.Segment{Flex: 1, BaseSize: 192, Widget: surfaceBox.Layout},
		block.Segment{Flex: 1, BaseSize: 192, Widget: surfaceVariantBox.Layout},
		block.Segment{Flex: 1, BaseSize: 192, Widget: outlineBox.Layout},
		block.Segment{Flex: 1, BaseSize: 192, Widget: outlineVariantBox.Layout},
	)

	return block.UniformPadding(examples.SpacingSmall).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return block.Line{
			Axis:     block.AxisHorizontal,
			Overflow: block.OverflowWrap,
			Expand:   true,
		}.Layout(gtx, widgets...)
	})
}
