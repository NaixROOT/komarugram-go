// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"

	"gio-mw/wdk"

	"gioui.org/layout"
)

const (
	jumpButtonSize = 40
	jumpIconSize   = 24
	jumpMargin     = 16
	jumpGap        = 12
	// historyPage is how many messages the store asks Telegram for at once.
	historyPage = 80
)

// atStart reports whether the history shows the very first message of the
// chat at its top: none is left to scroll to, loaded or not.
func (p *chatPage) atStart(history model.History) bool {
	pos := p.list.Position
	return pos.First == 0 && pos.Offset <= 0 && !history.HasOlder
}

// oneRequest reports whether the whole chat came in one request, which
// makes a button to its start no better than the scrollbar.
func (p *chatPage) oneRequest(history model.History) bool {
	return !history.HasOlder && !history.HasNewer && len(p.messages) < historyPage
}

// atEnd reports whether the history shows the newest message at its bottom.
func (p *chatPage) atEnd(history model.History) bool {
	return !p.list.Position.BeforeEnd && !history.HasNewer
}

// jump is where the history is to be put once a load that the store started
// for a jump has come.
type jump uint8

const (
	jumpNone jump = iota
	jumpStart
	jumpEnd
)

// revealed follows a store's start of a jump to an end of the history. A
// chat's history is opened again where the store says, and a thread's, which
// the store does not keep a viewport of, is put there when its load is over.
func (p *chatPage) revealed(to jump) {
	if p.threadRoot != 0 {
		p.jump = to
	} else {
		p.forget()
	}
	p.invalidate()
}

// finishJump puts the history at the end that a jump asked for, once what
// the store loads for it has come.
func (p *chatPage) finishJump(history model.History) {
	if p.jump == jumpNone || history.LoadingOlder || history.LoadingNewer {
		return
	}
	to := p.jump
	p.jump = jumpNone
	if history.Err != nil || len(p.messages) == 0 {
		return
	}
	if to == jumpStart {
		p.list.Position.First, p.list.Position.Offset = 0, 0
		p.list.Position.BeforeEnd = true
	} else {
		p.list.Position.BeforeEnd = false
	}
	p.invalidate()
}

// scrollToStart shows the chat's first message. When older messages are not
// loaded, the store loads the history again from the oldest one; a chat it
// cannot do that for, or one offline, scrolls to the top of what is loaded,
// and the loading of older messages goes on from there.
func (p *chatPage) scrollToStart(history model.History) {
	if history.HasOlder && !history.Offline {
		if r, ok := p.source.(model.HistoryEnds); ok && r.RevealFirst(p.chat) {
			p.revealed(jumpStart)
			return
		}
	}
	p.list.Position.First, p.list.Position.Offset = 0, 0
	p.list.Position.BeforeEnd = true
	p.invalidate()
}

// scrollToEnd shows the chat's newest message, loading the history again
// from it when the loaded part ends before it.
func (p *chatPage) scrollToEnd(history model.History) {
	if history.HasNewer && !history.Offline {
		if r, ok := p.source.(model.HistoryEnds); ok && r.RevealLast(p.chat) {
			p.revealed(jumpEnd)
			return
		}
	}
	p.list.Position.BeforeEnd = false
	p.invalidate()
}

// jumpButtons draws over the bottom right corner of the history, above
// end, the round buttons that scroll to its end and its start and retry a
// failed load, each there only when it has something to do, and above them
// the ring of a load on its way, of the size of a button.
func (p *chatPage) jumpButtons(gtx layout.Context, size image.Point, end int, history model.History, l localization.Catalog) {
	if !p.restored || len(p.messages) == 0 && history.Err == nil {
		return
	}
	if c := p.composer; c != nil && (c.pickerOpen || c.pickerDrawn) {
		return // the picker is over the corner they are in
	}
	type button struct {
		s     *surface
		icon  wdk.IconWidget
		label string
		show  bool
		do    func()
	}
	buttons := []button{
		{&p.toEnd, iconToBottom, l.T("history.to_end"), len(p.messages) > 0 && !p.atEnd(history), func() { p.scrollToEnd(history) }},
		{&p.toStart, iconToTop, l.T("history.to_top"), len(p.messages) > 0 && !p.atStart(history) && !p.oneRequest(history) && !history.LoadingOlder, func() { p.scrollToStart(history) }},
		{&p.retry, iconRefresh, l.T("history.retry"), history.Err != nil, func() {
			if s, ok := p.source.(interface{ Reload(int64) }); ok {
				s.Reload(p.chat)
			} else {
				p.source.LoadOlder(p.chat)
			}
		}},
	}
	sc := scheme(gtx)
	btn := gtx.Dp(jumpButtonSize)
	x := size.X - gtx.Dp(jumpMargin) - btn
	y := end - gtx.Dp(jumpMargin) - btn
	defer func() {
		if !(history.LoadingOlder || history.LoadingNewer) || len(p.messages) == 0 {
			return // the empty history shows its own ring
		}
		offset(gtx, image.Pt(x, y), func(gtx layout.Context) layout.Dimensions {
			fillRounded(gtx, sc.SurfaceContainerHigh, image.Pt(btn, btn), btn/2)
			px := gtx.Dp(jumpIconSize)
			return offset(gtx, image.Pt((btn-px)/2, (btn-px)/2), func(gtx layout.Context) layout.Dimensions {
				return exact(gtx, image.Pt(px, px), func(gtx layout.Context) layout.Dimensions { return p.loader.sized(gtx, l, jumpIconSize) })
			})
		})
	}()
	for _, b := range buttons {
		if !b.show {
			continue
		}
		if b.s.Clicked(gtx) {
			b.do()
			continue
		}
		offset(gtx, image.Pt(x, y), func(gtx layout.Context) layout.Dimensions {
			style := surfaceStyle{radius: btn / 2, background: sc.SurfaceContainerHigh, content: sc.SurfaceVariant.OnColor, button: b.label}
			return b.s.Layout(gtx, image.Pt(btn, btn), style, func(gtx layout.Context) layout.Dimensions {
				px := gtx.Dp(jumpIconSize)
				return offset(gtx, image.Pt((btn-px)/2, (btn-px)/2), func(gtx layout.Context) layout.Dimensions {
					return exact(gtx, image.Pt(px, px), func(gtx layout.Context) layout.Dimensions {
						return b.icon(gtx, sc.SurfaceVariant.OnColor)
					})
				})
			})
		})
		y -= btn + gtx.Dp(jumpGap)
	}
}
