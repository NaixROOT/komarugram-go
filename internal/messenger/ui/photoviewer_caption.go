// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"image/color"
	"math"
	"strings"

	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"

	"komarugram/internal/messenger/model"
	"komarugram/internal/messenger/styledtext"
)

// viewerCaption is the caption of the photo on screen, with its formatting
// and links; spoilers stay hidden.
type viewerCaption struct {
	key       model.MessageKey
	runs      []model.TextRun
	clusters  []styledtext.Cluster
	fragments []styledtext.Fragment
	click     gesture.Click
	// origin is where the text was drawn on the last frame.
	origin image.Point
}

// layoutCaption draws m's caption over the bottom of the stage, and returns
// where its top is.
func (v *photoViewer) layoutCaption(gtx layout.Context, m model.Message, stage image.Rectangle) int {
	c := &v.caption
	if c.key != m.Key {
		c.key, c.runs = m.Key, model.TextRuns(strings.TrimSpace(m.Text), m.Entities)
	}
	if len(c.runs) == 0 {
		return math.MaxInt
	}
	pad, padY, margin := gtx.Dp(12), gtx.Dp(8), gtx.Dp(12)
	for {
		e, ok := c.click.Update(gtx.Source)
		if !ok {
			break
		}
		if e.Kind != gesture.KindClick {
			continue
		}
		at := e.Position.Sub(image.Pt(pad, padY))
		for _, f := range c.fragments {
			if run := c.runs[f.Index]; run.URL != "" && at.In(f.Bounds) && v.openLink != nil {
				v.Close()
				v.openLink(run.URL)
				return math.MaxInt
			}
		}
	}
	theme := wdk.GetMaterialTheme(gtx)
	ty := theme.Typescale[token.TypestyleBodyLarge]
	white := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	styles := make([]styledtext.SpanStyle, len(c.runs))
	for i, run := range c.runs {
		styles[i] = runSpan(theme, ty, run, white, color.NRGBA{R: 0x8a, G: 0xc8, B: 0xff, A: 0xff}, white)
	}
	c.fragments = c.fragments[:0]
	text := styledtext.Text(theme.TextShaper, styles...)
	text.Clusters = &c.clusters
	text.Decorate = func(gtx layout.Context, f styledtext.Fragment, draw func()) {
		c.fragments = append(c.fragments, f)
		run, size := c.runs[f.Index], f.Bounds.Size()
		if run.Spoiler {
			paint.FillShape(gtx.Ops, color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff}, clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(3)).Op(gtx.Ops))
			return
		}
		draw()
		paintRunLine(gtx, run, size, styles[f.Index].Color)
	}
	limit := image.Pt(max(0, min(stage.Dx()-2*margin, gtx.Dp(640))-2*pad), stage.Dy()/3)
	tgtx := gtx
	tgtx.Constraints = layout.Constraints{Max: image.Pt(limit.X, math.MaxInt/2)}
	macro := op.Record(gtx.Ops)
	dims := text.Layout(tgtx, nil)
	call := macro.Stop()
	shown := image.Pt(dims.Size.X, min(dims.Size.Y, limit.Y))
	size := shown.Add(image.Pt(2*pad, 2*padY))
	at := image.Pt(stage.Min.X+(stage.Dx()-size.X)/2, stage.Max.Y-margin-size.Y)
	c.origin = at.Add(image.Pt(pad, padY))
	offset(gtx, at, func(gtx layout.Context) layout.Dimensions {
		paint.FillShape(gtx.Ops, color.NRGBA{A: 150}, clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(8)).Op(gtx.Ops))
		area := clip.Rect{Max: size}.Push(gtx.Ops)
		c.click.Add(gtx.Ops)
		for _, f := range c.fragments {
			if c.runs[f.Index].URL != "" {
				r := clip.Rect(f.Bounds.Add(image.Pt(pad, padY))).Push(gtx.Ops)
				pointer.CursorPointer.Add(gtx.Ops)
				r.Pop()
			}
		}
		area.Pop()
		offset(gtx, image.Pt(pad, padY), func(gtx layout.Context) layout.Dimensions {
			defer clip.Rect{Max: shown}.Push(gtx.Ops).Pop()
			call.Add(gtx.Ops)
			return dims
		})
		return layout.Dimensions{Size: size}
	})
	return at.Y
}
