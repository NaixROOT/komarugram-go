// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"time"

	"komarugram/internal/messenger/model"

	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
)

// sentNoteLife is how long the composer's note of a message sent waits for
// the message to show in the history: a send that failed leaves its note.
const sentNoteLife = 20 * time.Second

// sentNote is what the composer keeps of a message it sent, for the history
// to fly it up from the composer once it shows there.
type sentNote struct {
	chat int64
	at   time.Time
}

// noteSent notes that a message was sent to chat.
func (c *messageComposer) noteSent(chat int64) {
	c.sentNotes = append(c.sentNotes, sentNote{chat, time.Now()})
}

// forgetSent drops the note of a message to chat that was not sent.
func (c *messageComposer) forgetSent(chat int64) {
	for i, n := range c.sentNotes {
		if n.chat == chat {
			c.sentNotes = append(c.sentNotes[:i], c.sentNotes[i+1:]...)
			return
		}
	}
}

// takeSent reports whether a message was sent to chat, and drops its note.
func (c *messageComposer) takeSent(chat int64) bool {
	live := c.sentNotes[:0]
	for _, n := range c.sentNotes {
		if time.Since(n.at) < sentNoteLife {
			live = append(live, n)
		}
	}
	c.sentNotes = live
	for i, n := range c.sentNotes {
		if n.chat == chat {
			c.sentNotes = append(c.sentNotes[:i], c.sentNotes[i+1:]...)
			return true
		}
	}
	return false
}

// sendFlight is the flight of a message just sent, from the composer to its
// place in the history.
type sendFlight struct {
	tween wdk.FloatTween
}

// sentByUser reports whether the message is one that the user's own composer
// sent: an outgoing one, or any in Saved Messages, where Telegram does not
// mark the user's messages as outgoing.
func (p *chatPage) sentByUser(m model.Message) bool {
	return m.Outgoing || p.kind == model.KindSaved
}

// startFlights sets flying the messages of the user that came at the end of the
// history after prevLast, one for each message the composer sent. It does
// nothing with animations off.
func (p *chatPage) startFlights(gtx layout.Context, prevLast model.MessageID, animate bool) {
	if !animate || !wdk.AnimationsEnabled(gtx) || p.composer == nil || prevLast == 0 {
		return
	}
	for _, m := range p.messages {
		if m.Key.MessageID <= prevLast || !p.sentByUser(m) || !p.composer.takeSent(p.chat) {
			continue
		}
		f := &sendFlight{}
		f.tween.Duration = token.DurationMedium2
		f.tween.Easing = &token.EasingEmphasizedDecelerate
		f.tween.Animate(gtx, 0) // starts at the composer
		if p.flights == nil {
			p.flights = map[model.MessageID]*sendFlight{}
		}
		p.flights[m.Key.MessageID] = f
	}
}

// flyingRow draws a row with draw, which is a message that flies from the
// composer, if it does: it rises to its place while it fades in. from is how
// far below the row's middle the composer's is, in px.
func (p *chatPage) flyingRow(gtx layout.Context, id model.MessageID, from int, draw func(layout.Context) layout.Dimensions) layout.Dimensions {
	f := p.flights[id]
	if f == nil {
		return draw(gtx)
	}
	record := op.Record(gtx.Ops)
	dims := draw(gtx)
	call := record.Stop()
	t := f.tween.Animate(gtx, 1)
	if t >= 1 {
		delete(p.flights, id)
		call.Add(gtx.Ops)
		return dims
	}
	shift := int(float32(from+dims.Size.Y/2) * (1 - t))
	defer op.Offset(image.Pt(0, shift)).Push(gtx.Ops).Pop()
	defer paint.PushOpacity(gtx.Ops, t).Pop()
	call.Add(gtx.Ops)
	return dims
}
