// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"image"
	"strings"
	"time"
	"unicode/utf8"

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
	// asked is the query the results are for, once answered.
	asked      string
	answered   bool
	due        time.Time
	results    model.InlineResults
	generation int
	answers    chan inlineAnswer
	// more is set while the next page of answers is asked for.
	more   bool
	clicks []surface
	list   widget.List
	// area is where the answers were drawn on the last frame.
	area image.Rectangle
}

type inlineAnswer struct {
	generation int
	more       bool
	bot        *model.InlineBot
	query      string
	results    model.InlineResults
}

// updateInline follows the text of the field: a new bot is looked up, and a
// query goes once typing pauses.
func (c *messageComposer) updateInline(gtx layout.Context, chat int64, text string, permissions model.SendPermissions) {
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
			switch {
			case a.bot != nil:
				q.bot, q.botOK = *a.bot, a.bot.ID != 0
				q.due = gtx.Now
			case a.more:
				q.more = false
				if a.query == q.asked {
					q.results.Results = append(q.results.Results, a.results.Results...)
					q.results.Next = a.results.Next
				}
			default:
				q.results, q.asked, q.answered = a.results, a.query, true
				q.list.Position = layout.Position{}
			}
			continue
		default:
		}
		break
	}
	source, ok := c.source.(model.InlineBotSource)
	username, query, typed := model.InlineQuery(text)
	if !ok || !typed || !permissions.Allows(model.SendInline) {
		if q.username != "" {
			*q = inlineQuery{answers: q.answers, list: q.list}
			q.generation++
		}
		return
	}
	if !strings.EqualFold(username, q.username) {
		q.generation++
		// Not answered: even an empty query goes, as bots answer it too.
		q.username, q.query, q.botOK, q.results, q.answered = username, query, false, model.InlineResults{}, false
		generation, answers, invalidate := q.generation, q.answers, c.invalidate
		go func() {
			ctx, cancel := context.WithTimeout(c.ctx, 15*time.Second)
			defer cancel()
			bot, _ := source.InlineBot(ctx, username)
			answers <- inlineAnswer{generation: generation, bot: &bot}
			invalidate()
		}()
		return
	}
	if query != q.query {
		q.query, q.due = query, gtx.Now.Add(inlineDelay)
	}
	if !q.botOK || q.answered && q.query == q.asked || q.due.IsZero() {
		return
	}
	if gtx.Now.Before(q.due) {
		gtx.Execute(op.InvalidateCmd{At: q.due})
		return
	}
	q.due = time.Time{}
	c.askInline(source, chat, q.query, "")
}

// moreInline asks for the next page of answers.
func (c *messageComposer) moreInline(chat int64) {
	q := &c.inline
	source, ok := c.source.(model.InlineBotSource)
	if !ok || q.more || q.results.Next == "" || !q.botOK {
		return
	}
	q.more = true
	c.askInline(source, chat, q.asked, q.results.Next)
}

// askInline asks the bot for its answers to query from offset; a page past
// the first is more of them.
func (c *messageComposer) askInline(source model.InlineBotSource, chat int64, query, offset string) {
	q := &c.inline
	generation, answers, invalidate, bot := q.generation, q.answers, c.invalidate, q.bot.ID
	go func() {
		ctx, cancel := context.WithTimeout(c.ctx, 20*time.Second)
		defer cancel()
		results, _ := source.InlineResults(ctx, chat, bot, query, offset)
		answers <- inlineAnswer{generation: generation, more: offset != "", query: query, results: results}
		invalidate()
	}()
}

// layoutPlaceholder draws the bot's placeholder after "@bot " while the
// query is empty, as Telegram Desktop does.
func (q *inlineQuery) layoutPlaceholder(gtx layout.Context, length int) {
	text := "@" + q.username + " "
	if !q.botOK || q.bot.Placeholder == "" || q.query != "" || length != utf8.RuneCountInString(text) {
		return
	}
	sc := scheme(gtx)
	gtx.Constraints.Min = image.Point{}
	macro := op.Record(gtx.Ops)
	dims := label(gtx, text, token.TypestyleBodyLarge, sc.Surface.OnColor, 1)
	macro.Stop()
	offset(gtx, image.Pt(dims.Size.X, 0), func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = max(0, gtx.Constraints.Max.X-dims.Size.X)
		return label(gtx, q.bot.Placeholder, token.TypestyleBodyLarge, sc.SurfaceVariant.OnColor, 1)
	})
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
	gap := gtx.Dp(4)
	columns := max(1, (w-gap)/(gtx.Dp(96)+gap))
	cell := (w - gap*(columns+1)) / columns
	if q.results.Gallery {
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
		defer func() {
			if pos := q.list.Position; !pos.BeforeEnd {
				c.moreInline(chat)
			}
		}()
		if q.results.Gallery {
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
