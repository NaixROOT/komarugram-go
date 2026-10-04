// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"image"
	"strings"
	"time"

	"gio-mw/token"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/widget"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

// inlineDelay is how long typing pauses before a query goes, as Telegram
// Desktop's kInlineBotRequestDelay.
const inlineDelay = 400 * time.Millisecond

// inlineQuery is "@bot query" typed in the field and what the bot answered.
type inlineQuery struct {
	username, query string
	bot             model.InlineBot
	botOK           bool
	asked           string // the query the results are for
	due             time.Time
	results         model.InlineResults
	generation      int
	answers         chan inlineAnswer
	clicks          []surface
	list            widget.List
	// area is where the answers were drawn on the last frame.
	area image.Rectangle
}

type inlineAnswer struct {
	generation int
	username   string
	bot        *model.InlineBot
	query      string
	results    model.InlineResults
}

// updateInline follows the text of the field: a new bot is looked up, and a
// query goes once typing pauses.
func (c *messageComposer) updateInline(gtx layout.Context, chat int64, text string) {
	q := &c.inline
	if q.answers == nil {
		q.answers = make(chan inlineAnswer, 4)
		q.list.Axis = layout.Vertical
	}
	for {
		select {
		case a := <-q.answers:
			if a.generation != q.generation {
				continue
			}
			if a.bot != nil {
				q.bot, q.botOK = *a.bot, a.bot.ID != 0
				q.due = gtx.Now
			} else {
				q.results, q.asked = a.results, a.query
				q.list.Position = layout.Position{}
			}
			continue
		default:
		}
		break
	}
	source, ok := c.source.(model.InlineBotSource)
	username, query, typed := model.InlineQuery(text)
	if !ok || !typed || !c.permissions(chat).Allows(model.SendInline) {
		if q.username != "" {
			*q = inlineQuery{answers: q.answers, list: q.list}
			q.generation++
		}
		return
	}
	if !strings.EqualFold(username, q.username) {
		q.generation++
		q.username, q.query, q.botOK, q.results, q.asked = username, query, false, model.InlineResults{}, ""
		generation, answers, invalidate := q.generation, q.answers, c.invalidate
		go func() {
			ctx, cancel := context.WithTimeout(c.ctx, 15*time.Second)
			defer cancel()
			bot, _ := source.InlineBot(ctx, username)
			answers <- inlineAnswer{generation: generation, username: username, bot: &bot}
			invalidate()
		}()
		return
	}
	if query != q.query {
		q.query, q.due = query, gtx.Now.Add(inlineDelay)
	}
	if !q.botOK || q.query == q.asked || q.due.IsZero() {
		return
	}
	if gtx.Now.Before(q.due) {
		gtx.Execute(op.InvalidateCmd{At: q.due})
		return
	}
	q.due = time.Time{}
	generation, answers, invalidate, bot, query := q.generation, q.answers, c.invalidate, q.bot.ID, q.query
	go func() {
		ctx, cancel := context.WithTimeout(c.ctx, 20*time.Second)
		defer cancel()
		results, _ := source.InlineResults(ctx, chat, bot, query, "")
		answers <- inlineAnswer{generation: generation, query: query, results: results}
		invalidate()
	}()
}

// layoutInline draws the answers of the inline bot over the composer, as a
// grid of pictures for a gallery and as a list otherwise; a click sends one.
func (p *chatPage) layoutInline(gtx layout.Context, chat int64, size image.Point, pad, above int, l localization.Catalog, animate bool) {
	c := p.composer
	if c == nil {
		return
	}
	q := &c.inline
	results := q.results.Results
	if q.username == "" || len(results) == 0 {
		return
	}
	for len(q.clicks) < len(results) {
		q.clicks = append(q.clicks, surface{})
	}
	for i, r := range results {
		if q.clicks[i].Clicked(gtx) {
			d := c.draft(chat)
			item := r.Item
			c.submit(chat, model.OutgoingMessage{Item: &item})
			if d.err == nil {
				d.editor.SetText("")
				d.text = ""
			}
			return
		}
	}
	w := min(gtx.Dp(420), size.X-2*pad)
	content := len(results) * gtx.Dp(56)
	if q.results.Gallery {
		gap := gtx.Dp(4)
		columns := max(1, (w-gap)/(gtx.Dp(96)+gap))
		cell := (w - gap*(columns+1)) / columns
		content = (len(results)+columns-1)/columns*(cell+gap) + gap
	}
	h := min(content, gtx.Dp(300), above-gtx.Dp(16))
	if h <= 0 {
		return
	}
	rect := image.Rect(pad, above-gtx.Dp(8)-h, pad+w, above-gtx.Dp(8))
	q.area = rect
	inRect(gtx, rect, func(gtx layout.Context) layout.Dimensions {
		sc := scheme(gtx)
		menu := gtx.Constraints.Max
		radius := gtx.Dp(12)
		defer clip.UniformRRect(image.Rectangle{Max: menu}, radius).Push(gtx.Ops).Pop()
		overlayFill(gtx, p.menuBackdrop(), menu, rect.Min, sc.SurfaceContainerHigh, radius)
		if q.results.Gallery {
			cell, gap := gtx.Dp(96), gtx.Dp(4)
			columns := max(1, (menu.X-gap)/(cell+gap))
			cell = (menu.X - gap*(columns+1)) / columns
			rows := (len(results) + columns - 1) / columns
			return q.list.Layout(gtx, rows, func(gtx layout.Context, row int) layout.Dimensions {
				for col := range columns {
					i := row*columns + col
					if i >= len(results) {
						break
					}
					offset(gtx, image.Pt(gap+col*(cell+gap), gap), func(gtx layout.Context) layout.Dimensions {
						size := image.Pt(cell, cell)
						style := surfaceStyle{radius: gtx.Dp(6), background: sc.Surface.OnColor.SetOpacity(0), content: sc.Surface.OnColor, button: results[i].Title}
						return q.clicks[i].Layout(gtx, size, style, func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints = layout.Exact(size)
							return p.inlinePicture(gtx, results[i].Item.Media, size, animate, true)
						})
					})
				}
				return layout.Dimensions{Size: image.Pt(menu.X, cell+gap)}
			})
		}
		row := gtx.Dp(56)
		return q.list.Layout(gtx, len(results), func(gtx layout.Context, i int) layout.Dimensions {
			r := results[i]
			size := image.Pt(menu.X, row)
			style := surfaceStyle{background: sc.Surface.OnColor.SetOpacity(0), content: sc.Surface.OnColor, button: r.Title}
			return q.clicks[i].Layout(gtx, size, style, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints = layout.Exact(size)
				return layout.Inset{Left: 12, Right: 12}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							side := image.Pt(gtx.Dp(40), gtx.Dp(40))
							gtx.Constraints = layout.Exact(side)
							thumb := model.Message{Kind: model.MessagePhoto, Media: r.Thumb}
							return p.inlinePicture(gtx, thumb, side, false, true)
						}),
						layout.Rigid(layout.Spacer{Width: 12}.Layout),
						layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
							gtx.Constraints.Min = image.Point{}
							title := r.Title
							if title == "" {
								title = r.Kind
							}
							return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return label(gtx, title, token.TypestyleBodyMediumEmphasized, sc.Surface.OnColor, 1)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									return label(gtx, r.Description, token.TypestyleBodySmall, sc.SurfaceVariant.OnColor, 1)
								}),
							)
						}),
					)
				})
			})
		})
	})
}

// inlinePicture draws a result's picture over a plate the size of the cell.
func (p *chatPage) inlinePicture(gtx layout.Context, m model.Message, size image.Point, animate, crop bool) layout.Dimensions {
	defer clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(6)).Push(gtx.Ops).Pop()
	fillRect(gtx, scheme(gtx).SurfaceContainerHighest, size)
	if m.Media == nil || p.media == nil || p.images == nil {
		return layout.Dimensions{Size: size}
	}
	status := p.media.StatusFit(m, animate, size, crop)
	im := status.Frame
	if im == nil {
		im = status.Preview
	}
	if im != nil {
		widget.Image{Src: p.images.Op(im), Fit: widget.Cover, Position: layout.Center}.Layout(gtx)
	}
	return layout.Dimensions{Size: size}
}
