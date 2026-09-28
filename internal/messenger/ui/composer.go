// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"image"
	"reflect"
	"strings"
	"time"
	"unicode/utf16"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"

	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/widget/scroll"

	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
)

type messageDraft struct {
	editor   widget.Editor
	entities []model.Entity
	text     string
	pending  *model.OutgoingMessage
	sending  bool
	err      error
	// reply is the message what is sent next replies to, nil for none.
	reply *model.Message
}
type composerResult struct {
	chat    int64
	request model.OutgoingMessage
	err     error
}
type pickerResult struct {
	generation uint64
	page       model.PickerPage
	err        error
	append     bool
}
type featuredPackResult struct {
	generation uint64
	id         int64
	tab        model.PickerTab
	pack       model.StickerSet
	err        error
}
type messageComposer struct {
	editorHit              struct{}
	notice, lastNotice     string
	noticeUntil            time.Time
	source                 model.ComposerStore
	invalidate             func()
	ctx                    context.Context
	cancel                 context.CancelFunc
	drafts                 map[int64]*messageDraft
	chat                   int64
	attach, smile          surface
	dismiss                widget.Clickable
	send                   surface
	pickerOpen, attachOpen bool
	tab                    model.PickerTab
	search                 widget.Editor
	query                  string
	due                    time.Time
	generation             uint64
	loading                bool
	loadCancel             context.CancelFunc
	page                   model.PickerPage
	pickerErr              error
	results                chan pickerResult
	featuredResults        chan featuredPackResult
	featuredGeneration     uint64
	featuredCancel         context.CancelFunc
	featuredLoading        bool
	sends                  chan composerResult
	list                   scroll.List
	loader                 loadingIndicator
	strip                  layout.List
	hover                  hoverPlay
	// pickerDrawn is whether the picker was drawn in the last frame.
	pickerDrawn       bool
	stripDrag         stripDrag
	packClicks        map[int64]*surface
	itemClicks        map[string]*surface
	selectedPack      int64
	recent            [3][]model.PickerItem
	more, retry       surface
	attachmentActions [3]surface
	// replies is the strip of the message a draft replies to.
	replies replyBar
	// The picker, the attachment menu and the attachment forms open as
	// context menus; formShown is the form shown, kept while it closes.
	pickerMenu, attachMenu, formMenu contextMenu
	formShown                        int
	// tabs are the picker's tabs.
	tabs                        tabRow
	form                        int
	path, title, tasks          widget.Editor
	browse, confirm, cancelForm surface
	fileResults                 chan fileChoice
	choosing                    bool
}

func newMessageComposer(source model.ConversationStore, invalidate func()) *messageComposer {
	ctx, cancel := context.WithCancel(context.Background())
	c := &messageComposer{invalidate: invalidate, ctx: ctx, cancel: cancel, drafts: map[int64]*messageDraft{}, results: make(chan pickerResult, 8), featuredResults: make(chan featuredPackResult, 8), sends: make(chan composerResult, 8), fileResults: make(chan fileChoice, 1), packClicks: map[int64]*surface{}, itemClicks: map[string]*surface{}}
	c.source, _ = source.(model.ComposerStore)
	c.search.SingleLine = true
	c.path.SingleLine = true
	c.title.SingleLine = true
	c.list.Axis = layout.Vertical
	c.strip.Axis = layout.Horizontal

	return c
}
func (c *messageComposer) draft(chat int64) *messageDraft {
	d := c.drafts[chat]
	if d == nil {
		d = &messageDraft{editor: widget.Editor{Submit: true}}
		c.drafts[chat] = d
	}
	return d
}
func (c *messageComposer) request(l localization.Catalog, appendPage bool) {
	if !appendPage {
		c.cancelFeatured()
	}
	if c.loadCancel != nil {
		c.loadCancel()
	}
	c.generation++
	generation := c.generation
	c.loading = true
	c.pickerErr = nil
	ctx, cancel := context.WithTimeout(c.ctx, 90*time.Second)
	c.loadCancel = cancel
	req := model.PickerRequest{Tab: c.tab, Query: c.query, Language: string(l.Language()), ChatID: c.chat}
	if appendPage {
		req.Offset = c.page.Next
	}
	go func() {
		defer cancel()
		page, err := c.source.Picker(ctx, req)
		select {
		case c.results <- pickerResult{generation: generation, page: page, err: err, append: appendPage}:
			c.invalidate()
		case <-c.ctx.Done():
		}
	}()
}
func (c *messageComposer) cancelFeatured() {
	c.featuredGeneration++
	if c.featuredCancel != nil {
		c.featuredCancel()
		c.featuredCancel = nil
	}
	c.featuredLoading = false
}
func (c *messageComposer) selectedFeatured() *model.PickerPack {
	for i := range c.page.Featured {
		if c.page.Featured[i].ID == c.selectedPack {
			return &c.page.Featured[i]
		}
	}
	return nil
}
func (c *messageComposer) loadFeatured(pack model.PickerPack) {
	c.cancelFeatured()
	source, ok := c.source.(model.StickerSetStore)
	if !ok {
		c.pickerErr = fmt.Errorf("sticker set unavailable")
		return
	}
	ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
	c.featuredCancel = cancel
	c.featuredLoading = true
	c.pickerErr = nil
	generation, tab := c.featuredGeneration, c.tab
	go func() {
		defer cancel()
		full, err := source.StickerSet(ctx, pack.Ref)
		select {
		case c.featuredResults <- featuredPackResult{generation: generation, id: pack.ID, tab: tab, pack: full, err: err}:
			c.invalidate()
		case <-c.ctx.Done():
		}
	}()
}
func randomMessageID() int64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UnixNano()
	}
	v := int64(binary.LittleEndian.Uint64(b[:]))
	if v == 0 {
		return 1
	}
	return v
}
func (c *messageComposer) submit(chat int64, msg model.OutgoingMessage) {
	d := c.draft(chat)
	if d.sending || c.source == nil {
		return
	}
	if d.reply != nil && msg.ReplyTo == 0 {
		msg.ReplyTo = d.reply.Key.MessageID
	}
	if msg.RandomID == 0 {
		if d.pending != nil {
			prior := *d.pending
			id := prior.RandomID
			prior.RandomID = 0
			if reflect.DeepEqual(prior, msg) {
				msg.RandomID = id
			}
		}
		if msg.RandomID == 0 {
			msg.RandomID = randomMessageID()
		}
	}
	if err := msg.Validate(); err != nil {
		d.err = err
		return
	}
	d.text = d.editor.Text()
	d.pending = &msg
	d.sending = true
	d.err = nil
	go func() {
		ctx, cancel := context.WithTimeout(c.ctx, 10*time.Minute)
		defer cancel()
		err := c.source.Send(ctx, chat, msg)
		select {
		case c.sends <- composerResult{chat: chat, request: msg, err: err}:
			c.invalidate()
		case <-c.ctx.Done():
		}
	}()
}
func (c *messageComposer) submitText() {
	d := c.draft(c.chat)
	if d.pending != nil && d.err != nil {
		c.submit(c.chat, *d.pending)
		return
	}
	if strings.TrimSpace(d.editor.Text()) == "" {
		return
	}
	c.submit(c.chat, model.OutgoingMessage{Text: d.editor.Text(), Entities: append([]model.Entity(nil), d.entities...)})
}
func (c *messageComposer) update(gtx layout.Context, chat int64, l localization.Catalog) {
	if c.chat != chat {
		c.cancelFeatured()
		c.chat = chat
		c.pickerOpen = false
		c.attachOpen = false
		c.form = 0
	}
	for {
		select {
		case r := <-c.sends:
			d := c.draft(r.chat)
			d.sending = false
			d.err = r.err
			if r.err == nil {
				if r.request.Item == nil && r.request.Path == "" && len(r.request.Tasks) == 0 && d.editor.Text() == r.request.Text {
					d.editor.SetText("")
					d.text = ""
					d.entities = nil
				}
				d.pending = nil
				if d.reply != nil && r.request.ReplyTo == d.reply.Key.MessageID {
					d.reply = nil
				}
				if r.chat == c.chat {
					c.form = 0
				}
			}
		case r := <-c.results:
			if r.generation != c.generation {
				continue
			}
			c.loading = false
			c.pickerErr = r.err
			if r.err == nil {
				if r.append {
					c.page.Items = model.PreferSaved(c.page.Items, r.page.Items)
					c.page.Next = r.page.Next
				} else {
					c.page = r.page
					c.itemClicks = map[string]*surface{}
					if c.query == "" {
						if pack := c.selectedFeatured(); pack != nil && !pack.Complete {
							c.loadFeatured(*pack)
						}
					}
				}
			}
		case r := <-c.featuredResults:
			if r.generation != c.featuredGeneration || r.tab != c.tab || r.id != c.selectedPack {
				continue
			}
			c.featuredLoading = false
			c.featuredCancel = nil
			c.pickerErr = r.err
			if r.err == nil {
				if pack := c.selectedFeatured(); pack != nil {
					pack.Items = r.pack.Items
					pack.Title = r.pack.Title
					pack.Ref = r.pack.Ref
					pack.Complete = true
					c.list.Position = layout.Position{}
				}
			}
		case f := <-c.fileResults:
			c.choosing = false
			if f.chat != c.chat || f.form != c.form {
				continue
			}
			if f.err != nil {
				c.draft(c.chat).err = f.err
			} else if f.path != "" {
				c.path.SetText(f.path)
			}
		default:
			goto drained
		}
	}
drained:
	d := c.draft(chat)
	for {
		_, ok := gtx.Event(pointer.Filter{Target: &c.editorHit, Kinds: pointer.Press})
		if !ok {
			break
		}
		if !d.sending && c.source != nil {
			gtx.Execute(key.FocusCmd{Tag: &d.editor})
		}
		c.attachOpen = false
	}
	for i := range c.attachmentActions {
		if c.attachmentActions[i].Clicked(gtx) {
			c.form = i + 1
			c.attachOpen = false
			c.pickerOpen = false
			editor := &c.path
			if c.form == 3 {
				editor = &c.title
			}
			gtx.Execute(key.FocusCmd{Tag: editor})
		}
	}

	if !d.sending {
		for {
			ev, ok := d.editor.Update(gtx)
			if !ok {
				break
			}
			switch ev.(type) {
			case widget.SubmitEvent:
				c.submitText()
			case widget.ChangeEvent:
				next := d.editor.Text()
				if next == d.text {
					continue
				}
				d.entities = shiftEmojiEntities(d.text, next, d.entities)
				d.text = next
				d.pending = nil
				d.err = nil
				if g, ok := c.source.(model.GhostStore); ok && strings.TrimSpace(next) != "" {
					g.Typing(c.chat)
				}
			}
		}
	}
	for {
		_, ok := c.search.Update(gtx)
		if !ok {
			break
		}
	}
	q := strings.TrimSpace(c.search.Text())
	if q != c.query {
		c.cancelFeatured()
		c.query = q
		c.due = gtx.Now.Add(300 * time.Millisecond)
		c.generation++
		if c.loadCancel != nil {
			c.loadCancel()
		}
		c.loading = false
		c.page.Items = nil
		c.page.Next = ""
		c.selectedPack = 0
		c.list.Position = layout.Position{}
	}
	if c.smile.Clicked(gtx) {
		c.pickerOpen = !c.pickerOpen
		if c.pickerOpen {
			gtx.Execute(key.FocusCmd{Tag: &c.search})
		}
		c.attachOpen = false
		c.form = 0
		if c.pickerOpen && c.source != nil {
			c.request(l, false)
		}
	}
	if c.attach.Clicked(gtx) {
		c.attachOpen = !c.attachOpen
		if c.attachOpen {
			gtx.Execute(key.FocusCmd{Tag: &d.editor})
		}
		c.pickerOpen = false
		c.form = 0
	}
	if c.send.Clicked(gtx) {
		c.submitText()
	}
	if c.dismiss.Clicked(gtx) {
		c.pickerOpen = false
		c.attachOpen = false
		c.form = 0
	}
	if i, ok := c.tabs.Clicked(gtx, int(c.tab), 3); ok {
		c.cancelFeatured()
		c.tab = model.PickerTab(i)
		c.search.SetText("")
		c.query = ""
		c.due = time.Time{}
		c.page = model.PickerPage{}
		c.selectedPack = 0
		c.list.Position = layout.Position{}
		c.strip.Position = layout.Position{}
		if c.source != nil {
			c.request(l, false)
		}
	}
	for c.pickerOpen || c.attachOpen || c.form != 0 {
		ev, ok := gtx.Event(key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if e, ok := ev.(key.Event); ok && e.State == key.Press {
			c.pickerOpen = false
			c.attachOpen = false
			c.form = 0
			gtx.Execute(key.FocusCmd{Tag: &d.editor})
		}
	}
	if c.pickerOpen && !c.due.IsZero() {
		if !gtx.Now.Before(c.due) {
			c.due = time.Time{}
			if c.source != nil {
				c.request(l, false)
			}
		} else {
			gtx.Execute(op.InvalidateCmd{At: c.due})
		}
	}
	if !c.pickerOpen && c.featuredLoading {
		c.cancelFeatured()
	}
}

// Preserve custom-emoji UTF-16 entities across ordinary edits, dropping only overlaps.
func shiftEmojiEntities(old, new string, entities []model.Entity) []model.Entity {
	a, b := []rune(old), []rune(new)
	prefix := 0
	for prefix < len(a) && prefix < len(b) && a[prefix] == b[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < len(a)-prefix && suffix < len(b)-prefix && a[len(a)-1-suffix] == b[len(b)-1-suffix] {
		suffix++
	}
	start := len(utf16.Encode(a[:prefix]))
	end := len(utf16.Encode(a[:len(a)-suffix]))
	delta := len(utf16.Encode(b)) - len(utf16.Encode(a))
	var out []model.Entity
	for _, e := range entities {
		if e.Offset+e.Length <= start {
			out = append(out, e)
		} else if e.Offset >= end {
			e.Offset += delta
			out = append(out, e)
		}
	}
	return out
}
func flatEditor(gtx layout.Context, e *widget.Editor, hint string) layout.Dimensions {
	theme := wdk.GetMaterialTheme(gtx)
	sc := scheme(gtx)
	style := theme.Typescale[token.TypestyleBodyLarge]
	text := op.Record(gtx.Ops)
	paint.ColorOp{Color: sc.Surface.OnColor.AsNRGBA()}.Add(gtx.Ops)
	textOp := text.Stop()
	selected := op.Record(gtx.Ops)
	paint.ColorOp{Color: sc.Primary.Color.SetOpacity(token.OpacityLevel3).AsNRGBA()}.Add(gtx.Ops)
	selection := selected.Stop()
	if e.Text() == "" {
		label(gtx, hint, token.TypestyleBodyLarge, sc.SurfaceVariant.OnColor, 1)
	}
	e.LineHeight = style.LineHeight
	return e.Layout(gtx, theme.TextShaper, style.AsRegularFont(), style.Size, textOp, selection)
}

const (
	// composerBlurRadius is how much the history behind the floating
	// composer is blurred.
	composerBlurRadius = unit.Dp(16)
	// composerBlurOpacity is the opacity of the composer over the blur.
	composerBlurOpacity = token.OpacityLevel(0.7)
)

// layoutBackdrop fills the current clip, of size, with backdrop blurred:
// the history, whose origin is at -origin.
func layoutBackdrop(gtx layout.Context, size, origin image.Point, backdrop op.CallOp) {
	radius := gtx.Dp(composerBlurRadius)
	defer paint.PushBlur(gtx.Ops, float32(radius)).Pop()
	// The blur reaches past the clip by about radius: cover that much, the
	// window's edge included, so that no transparency is blurred in.
	margin := image.Pt(2*radius, 2*radius)
	defer clip.Rect{Min: margin.Mul(-1), Max: size.Add(margin)}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: scheme(gtx).SurfaceContainerLow.AsNRGBA()}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	defer op.Offset(origin.Mul(-1)).Push(gtx.Ops).Pop()
	backdrop.Add(gtx.Ops)
}

// composerBarHeight is the height of the composer's bar in a chat of size.
func composerBarHeight(gtx layout.Context, size image.Point) int {
	return min(gtx.Dp(56), size.Y)
}

// floatingComposerCover is how much of the bottom of a chat of size the
// floating composer covers, with the gap below it.
func floatingComposerCover(gtx layout.Context, size image.Point) int {
	height := composerBarHeight(gtx, size)
	return height + min(gtx.Dp(12), max(0, size.Y-height))
}

// Layout draws the composer over the chat page p. A floating composer shows
// backdrop, the history behind it, blurred; nil draws it opaque.
func (c *messageComposer) Layout(gtx layout.Context, chat int64, l localization.Catalog, p *chatPage, animate bool, backdrop *op.CallOp) layout.Dimensions {
	c.update(gtx, chat, l)
	p.errorMu.Lock()
	mediaErr := p.mediaError
	p.errorMu.Unlock()
	notice := p.selectionNotice
	if mediaErr != nil {
		notice = mediaErrorText(mediaErr)
	}
	if notice != c.lastNotice {
		c.lastNotice = notice
		c.notice = notice
		c.noticeUntil = gtx.Now.Add(5 * time.Second)
	}
	size := gtx.Constraints.Max
	pad := min(gtx.Dp(16), size.X/8)
	classic := p.classicComposer()
	height := composerBarHeight(gtx, size)
	bottom := floatingComposerCover(gtx, size) - height
	rect := image.Rect(pad, max(0, size.Y-bottom-height), max(pad, size.X-pad), size.Y-bottom)
	if classic {
		rect = image.Rect(0, max(0, size.Y-height), size.X, size.Y)
	}
	if p.frozen.Frozen() {
		// A frozen account cannot send: the bar tells why, as Telegram
		// Desktop's does.
		inRect(gtx, rect, func(gtx layout.Context) layout.Dimensions {
			sc := scheme(gtx)
			s := gtx.Constraints.Max
			radius := s.Y / 2
			if classic {
				radius = 0
			}
			defer clip.UniformRRect(image.Rectangle{Max: s}, radius).Push(gtx.Ops).Pop()
			fill := sc.SurfaceContainerHigh
			if backdrop != nil {
				layoutBackdrop(gtx, s, rect.Min, *backdrop)
				fill = fill.SetOpacity(composerBlurOpacity)
			}
			fillRounded(gtx, fill, s, radius)
			if classic {
				fillRect(gtx, sc.OutlineVariant, image.Pt(s.X, gtx.Dp(1)))
			}
			return p.frozen.layoutComposer(gtx, l, s, radius)
		})
		return layout.Dimensions{Size: size}
	}
	d := c.draft(chat)
	if c.pickerOpen || c.attachOpen || c.form != 0 {
		c.dismiss.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return layout.Dimensions{Size: size} })
	}
	// above is the top of the composer, the reply strip included: what
	// opens from the composer opens over it.
	c.replyLayout(gtx, chat, rect, classic, backdrop, l, p)
	above := rect.Min.Y - c.replyHeight(gtx, chat, classic)
	inRect(gtx, rect, func(gtx layout.Context) layout.Dimensions {
		sc := scheme(gtx)
		s := gtx.Constraints.Max
		radius := s.Y / 2
		if classic {
			radius = 0
		}
		// Keep pointer input inside the composer from reaching the history.
		defer clip.UniformRRect(image.Rectangle{Max: s}, radius).Push(gtx.Ops).Pop()
		fill := sc.SurfaceContainerHigh
		if backdrop != nil {
			layoutBackdrop(gtx, s, rect.Min, *backdrop)
			fill = fill.SetOpacity(composerBlurOpacity)
		}
		fillRounded(gtx, fill, s, radius)
		if classic {
			fillRect(gtx, sc.OutlineVariant, image.Pt(s.X, gtx.Dp(1)))
		}
		event.Op(gtx.Ops, c)
		// Square slots put the icons' circles at the centers of the
		// capsule's rounded ends, as far from its edges across as along.
		iconWidth := min(s.Y, s.X/3)
		button := func(x int, click *surface, icon wdk.IconWidget, label string) {
			inRect(gtx, image.Rect(x, 0, x+iconWidth, s.Y), func(gtx layout.Context) layout.Dimensions {
				size := gtx.Constraints.Max
				// The whole slot takes clicks; the state layer and the ripple
				// are a circle around the icon, as on Material icon buttons.
				d := min(gtx.Dp(40), size.X, size.Y)
				circle := image.Rectangle{Max: image.Pt(d, d)}.Add(size.Sub(image.Pt(d, d)).Div(2))
				content := sc.SurfaceVariant.OnColor
				style := surfaceStyle{area: circle, radius: d / 2, background: content.SetOpacity(0), content: content, button: label}
				return click.Layout(gtx, size, style, func(gtx layout.Context) layout.Dimensions {
					return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return icon(gtx, content) })
				})
			})
		}
		button(0, &c.attach, iconAttach, l.T("composer.attach"))
		button(s.X-iconWidth, &c.smile, iconEmoji, l.T("composer.emoji"))
		sendWidth := 0
		if strings.TrimSpace(d.editor.Text()) != "" || d.sending || d.pending != nil {
			sendWidth = min(gtx.Dp(92), s.X/3)
			inRect(gtx, image.Rect(s.X-iconWidth-sendWidth, 0, s.X-iconWidth, s.Y), func(gtx layout.Context) layout.Dimensions {
				if d.sending {
					gtx = gtx.Disabled()
				}
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					txt := l.T("composer.send")
					if d.sending {
						txt = l.T("composer.sending")
					}
					if d.err != nil && d.pending != nil {
						txt = l.T("composer.retry")
					}
					return textButton(gtx, &c.send, txt)
				})
			})
		}
		inRect(gtx, image.Rect(iconWidth, 0, max(iconWidth, s.X-iconWidth-sendWidth), s.Y), func(gtx layout.Context) layout.Dimensions {
			if d.sending || c.source == nil {
				gtx = gtx.Disabled()
			}
			size := gtx.Constraints.Max
			inner := gtx
			inner.Constraints.Min.Y = 0
			inner.Constraints.Max.Y = max(0, size.Y-gtx.Dp(12))
			record := op.Record(gtx.Ops)
			dims := flatEditor(inner, &d.editor, l.T("composer.message"))
			call := record.Stop()
			transform := op.Offset(image.Pt(0, max(0, (size.Y-dims.Size.Y)/2))).Push(gtx.Ops)
			call.Add(gtx.Ops)
			transform.Pop()
			// Pass presses to the editor for caret placement, but focus it from the
			// capsule's full height, including the padding above and below the text.
			area := clip.Rect{Max: size}.Push(gtx.Ops)
			pass := pointer.PassOp{}.Push(gtx.Ops)
			event.Op(gtx.Ops, &c.editorHit)
			pass.Pop()
			area.Pop()
			return layout.Dimensions{Size: size}
		})
		return layout.Dimensions{Size: s}
	})
	if d.err != nil && c.form == 0 {
		inRect(gtx, image.Rect(pad, max(0, above-gtx.Dp(52)), size.X-pad, above-gtx.Dp(4)), func(gtx layout.Context) layout.Dimensions { return pill(gtx, mediaErrorText(d.err)) })
	}
	if c.notice != "" && gtx.Now.Before(c.noticeUntil) && d.err == nil && c.form == 0 && !c.pickerOpen && !c.attachOpen {
		inRect(gtx, image.Rect(pad, max(0, above-gtx.Dp(48)), size.X-pad, above-gtx.Dp(4)), func(gtx layout.Context) layout.Dimensions { return pill(gtx, c.notice) })
		gtx.Execute(op.InvalidateCmd{At: c.noticeUntil})
	}
	{
		// The picker opens from the emoji button, at the composer's end.
		w := min(gtx.Dp(372), max(0, size.X-2*pad))
		h := min(gtx.Dp(488), max(0, above-gtx.Dp(12)))
		area := image.Rect(size.X-pad-w, above-gtx.Dp(8)-h, size.X-pad, above-gtx.Dp(8))
		drawn := false
		c.pickerMenu.Layout(gtx, c.pickerOpen, area, menuFromBottomRight, gtx.Dp(16), func(gtx layout.Context) layout.Dimensions {
			drawn = true
			return c.pickerLayout(gtx, l, p, animate)
		})
		// The picker closed: its stickers keep their first frames only.
		if c.pickerDrawn && !drawn {
			p.dropStickerLoops()
		}
		if drawn && !c.pickerDrawn {
			p.stickerViewShown()
		}
		c.pickerDrawn = drawn
	}
	{
		// The attachment menu opens from the paperclip, at the start.
		w := min(gtx.Dp(264), size.X-2*pad)
		h := min(gtx.Dp(144), max(0, above-gtx.Dp(8)))
		area := image.Rect(pad, above-h-gtx.Dp(8), pad+w, above-gtx.Dp(8))
		c.attachMenu.Layout(gtx, c.attachOpen, area, menuFromBottomLeft, gtx.Dp(12), func(gtx layout.Context) layout.Dimensions {
			menuSize := gtx.Constraints.Max
			sc := scheme(gtx)
			defer clip.UniformRRect(image.Rectangle{Max: menuSize}, gtx.Dp(12)).Push(gtx.Ops).Pop()
			fillRounded(gtx, sc.SurfaceContainerHigh, menuSize, gtx.Dp(12))
			// Without this clip the menu's input region covers the surrounding chat
			// and prevents the outside-click handler from closing the menu.
			event.Op(gtx.Ops, &c.attachmentActions)
			icons := []wdk.IconWidget{iconAttachPhoto, iconAttachFile, iconAttachTasks}
			for i, k := range []string{"photo", "file", "tasks"} {
				inRect(gtx, image.Rect(0, i*h/3, w, (i+1)*h/3), func(gtx layout.Context) layout.Dimensions {
					row := gtx.Constraints.Max
					style := surfaceStyle{background: sc.Surface.OnColor.SetOpacity(0), content: sc.Surface.OnColor}
					return c.attachmentActions[i].Layout(gtx, row, style, func(gtx layout.Context) layout.Dimensions {
						if i > 0 {
							fillRect(gtx, sc.OutlineVariant, image.Pt(row.X, gtx.Dp(1)))
						}
						return layout.Inset{Left: 16, Right: 16}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
								layout.Rigid(func(gtx layout.Context) layout.Dimensions { return icons[i](gtx, sc.SurfaceVariant.OnColor) }),
								layout.Rigid(layout.Spacer{Width: 12}.Layout),
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return layout.W.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
										gtx.Constraints.Min = image.Point{}
										return label(gtx, l.T("composer."+k), token.TypestyleBodyMedium, sc.Surface.OnColor, 1)
									})
								}),
							)
						})
					})
				})
			}
			return layout.Dimensions{Size: menuSize}
		})
	}
	bar := rect
	bar.Min.Y = above
	if classic {
		// Keep the form off the window's edge, where the capsule would start.
		bar.Min.X, bar.Max.X = pad, max(pad, size.X-pad)
	}
	c.formLayout(gtx, bar, l)
	return layout.Dimensions{Size: size}
}
func (c *messageComposer) formLayout(gtx layout.Context, bar image.Rectangle, l localization.Catalog) {
	if c.form != 0 {
		c.formShown = c.form
	}
	form := c.formShown
	w := min(gtx.Dp(420), bar.Dx())
	h := min(gtx.Dp(330), max(0, bar.Min.Y-gtx.Dp(8)))
	area := image.Rect(bar.Min.X, bar.Min.Y-gtx.Dp(8)-h, bar.Min.X+w, bar.Min.Y-gtx.Dp(8))
	// A form opens from the attachment menu, at the composer's start.
	c.formMenu.Layout(gtx, c.form != 0, area, menuFromBottomLeft, gtx.Dp(16), func(gtx layout.Context) layout.Dimensions {
		s := gtx.Constraints.Max
		fillRounded(gtx, scheme(gtx).SurfaceContainerHigh, s, gtx.Dp(16))
		defer clip.Rect{Max: s}.Push(gtx.Ops).Pop()
		event.Op(gtx.Ops, c)
		if c.cancelForm.Clicked(gtx) {
			c.form = 0
		}
		if c.browse.Clicked(gtx) && !c.choosing {
			c.choosing = true
			ctx := c.ctx
			media := form == 1
			chat := c.chat
			go func() {
				f := chooseAttachment(ctx, media)
				f.chat, f.form = chat, form
				select {
				case c.fileResults <- f:
					c.invalidate()
				case <-ctx.Done():
				}
			}()
		}
		if c.confirm.Clicked(gtx) {
			msg := model.OutgoingMessage{Path: strings.TrimSpace(c.path.Text()), AsMedia: form == 1}
			if form == 3 {
				msg.Path = ""
				msg.Text = strings.TrimSpace(c.title.Text())
				for _, task := range strings.Split(c.tasks.Text(), "\n") {
					if task = strings.TrimSpace(task); task != "" {
						msg.Tasks = append(msg.Tasks, task)
					}
				}
				if len(msg.Tasks) == 0 {
					c.draft(c.chat).err = fmt.Errorf("%s", l.T("composer.items"))
				}
			}
			if form != 3 || len(msg.Tasks) > 0 {
				c.submit(c.chat, msg)
			}
		}
		return layout.UniformInset(16).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if c.draft(c.chat).sending {
				gtx = gtx.Disabled()
			}
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					k := "file"
					if form == 1 {
						k = "photo"
					}
					if form == 3 {
						k = "tasks"
					}
					return label(gtx, l.T("composer."+k), token.TypestyleTitleMedium, scheme(gtx).Surface.OnColor, 1)
				}), vspace(16),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					if err := c.draft(c.chat).err; err != nil {
						return label(gtx, mediaErrorText(err), token.TypestyleBodySmall, scheme(gtx).Error.Color, 3)
					}
					return layout.Dimensions{}
				}),
				layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
					if form != 3 {
						return flatEditor(gtx, &c.path, l.T("composer.path"))
					}
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions { return flatEditor(gtx, &c.title, l.T("composer.title")) }), vspace(16), layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return flatEditor(gtx, &c.tasks, l.T("composer.items")) }), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return label(gtx, l.T("composer.task_hint"), token.TypestyleBodySmall, scheme(gtx).SurfaceVariant.OnColor, 2)
					}))
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if form == 3 {
							return layout.Dimensions{}
						}
						return textButton(gtx, &c.browse, l.T("composer.browse"))
					}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return textButton(gtx, &c.cancelForm, l.T("composer.cancel"))
					}), layout.Rigid(func(gtx layout.Context) layout.Dimensions { return textButton(gtx, &c.confirm, l.T("composer.send")) }))
				}),
			)
		})
	})
}
