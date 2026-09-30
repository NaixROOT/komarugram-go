// SPDX-License-Identifier: Unlicense OR MIT

package exp

import (
	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// TODO: Implement surface wdk.Box variants.

// Background fills the application background, it's not for other widgets.
func Background(gtx layout.Context) {
	windowArea := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	materialTheme := wdk.GetMaterialTheme(gtx)
	// The background is the root surface of a frame: the frame's surfaces
	// start afresh on it, should it be drawn twice.
	surfaces(gtx).reset()
	NewSurfaceTheme(gtx, materialTheme.Scheme.Background)
	paint.Fill(gtx.Ops, materialTheme.Scheme.Background.Color.AsNRGBA())
	windowArea.Pop()
}

func PrimaryContainer(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, materialTheme.Scheme.PrimaryContainer)
	paint.FillShape(gtx.Ops, materialTheme.Scheme.PrimaryContainer.Color.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func SecondaryContainer(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, materialTheme.Scheme.SecondaryContainer)
	paint.FillShape(gtx.Ops, materialTheme.Scheme.SecondaryContainer.Color.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func TertiaryContainer(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, materialTheme.Scheme.TertiaryContainer)
	paint.FillShape(gtx.Ops, materialTheme.Scheme.TertiaryContainer.Color.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func InverseSurface(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, materialTheme.Scheme.InverseSurface)
	paint.FillShape(gtx.Ops, materialTheme.Scheme.InverseSurface.Color.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func Surface(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, materialTheme.Scheme.Surface)
	paint.FillShape(gtx.Ops, materialTheme.Scheme.Surface.Color.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func SurfaceDim(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, token.MatColorSet{
		Color:   materialTheme.Scheme.SurfaceDim,
		OnColor: materialTheme.Scheme.Surface.Color,
	})
	paint.FillShape(gtx.Ops, materialTheme.Scheme.SurfaceDim.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func SurfaceBright(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, token.MatColorSet{
		Color:   materialTheme.Scheme.SurfaceBright,
		OnColor: materialTheme.Scheme.Surface.Color,
	})
	paint.FillShape(gtx.Ops, materialTheme.Scheme.SurfaceBright.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func SurfaceVariant(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, materialTheme.Scheme.SurfaceVariant)
	paint.FillShape(gtx.Ops, materialTheme.Scheme.SurfaceVariant.Color.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func SurfaceContainerLowest(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, token.MatColorSet{
		Color:   materialTheme.Scheme.SurfaceContainerLowest,
		OnColor: materialTheme.Scheme.Surface.Color,
	})
	paint.FillShape(gtx.Ops, materialTheme.Scheme.SurfaceContainerLowest.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func SurfaceContainerLow(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, token.MatColorSet{
		Color:   materialTheme.Scheme.SurfaceContainerLow,
		OnColor: materialTheme.Scheme.Surface.Color,
	})
	paint.FillShape(gtx.Ops, materialTheme.Scheme.SurfaceContainerLow.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func SurfaceContainer(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, token.MatColorSet{
		Color:   materialTheme.Scheme.SurfaceContainer,
		OnColor: materialTheme.Scheme.Surface.Color,
	})
	paint.FillShape(gtx.Ops, materialTheme.Scheme.SurfaceContainer.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func SurfaceContainerHigh(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, token.MatColorSet{
		Color:   materialTheme.Scheme.SurfaceContainerHigh,
		OnColor: materialTheme.Scheme.Surface.Color,
	})
	paint.FillShape(gtx.Ops, materialTheme.Scheme.SurfaceContainerHigh.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func SurfaceContainerHighest(gtx layout.Context, shapeArea clip.Op) *SurfaceTheme {
	materialTheme := wdk.GetMaterialTheme(gtx)
	surfaceTheme := NewSurfaceTheme(gtx, token.MatColorSet{
		Color:   materialTheme.Scheme.SurfaceContainerHighest,
		OnColor: materialTheme.Scheme.Surface.Color,
	})
	paint.FillShape(gtx.Ops, materialTheme.Scheme.SurfaceContainerHighest.AsNRGBA(), shapeArea)
	return surfaceTheme
}

func DisplayL(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleDisplayLarge,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func DisplayM(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleDisplayMedium,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func DisplayS(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleDisplaySmall,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func HeadlineL(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleHeadlineLarge,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func HeadlineM(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleHeadlineMedium,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func HeadlineS(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleHeadlineSmall,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func TitleL(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleTitleLarge,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func TitleM(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleTitleMedium,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func TitleS(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleTitleSmall,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func BodyL(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleBodyLarge,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func BodyM(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleBodyMedium,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func BodyS(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleBodySmall,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func LabelL(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleLabelLarge,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func LabelLP(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleLabelLarge,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func LabelM(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleLabelMedium,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func LabelMP(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleLabelMedium,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func LabelS(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestyleLabelSmall,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}

func Preformatted(gtx layout.Context, txt string) layout.Dimensions {
	surfaceTheme := GetSurfaceTheme(gtx)
	presentation := wdk.LabelStyle{
		Color:     surfaceTheme.OnColor,
		Typestyle: token.TypestylePreformatted,
	}
	return wdk.LayoutLabel(gtx, presentation, txt)
}
