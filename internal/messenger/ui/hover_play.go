// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"time"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
)

// How long a hovered item waits before it plays: after the pointer last
// moved, and after a list of the view last scrolled.
const (
	hoverMoveDelay   = 100 * time.Millisecond
	hoverScrollDelay = 250 * time.Millisecond
)

// hoverPlay decides which item of a scrolling grid plays on hover when
// automatic animations are off. An item plays once the pointer has rested on
// it for hoverMoveDelay and no list of the view has scrolled for
// hoverScrollDelay, so sweeping the pointer over a grid or scrolling it
// under the pointer starts no decoder that would be stopped at once. An item
// cheap to play, which starts no costly decoder, plays at once. The playing
// item keeps playing while hovered, the pointer moving or not.
type hoverPlay struct {
	key any
	// due is when the hovered item may play.
	due     time.Time
	playing bool
	lists   []*layout.List
	offsets []layout.Position
}

// Update reads the pointer's moves over the view and remembers its lists;
// call it at the top of the view's layout.
func (h *hoverPlay) Update(gtx layout.Context, lists ...*layout.List) {
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: h, Kinds: pointer.Move | pointer.Enter | pointer.Leave})
		if !ok {
			break
		}
		if _, ok := ev.(pointer.Event); ok && !h.playing {
			h.delay(gtx, hoverMoveDelay)
		}
	}
	h.lists = lists
}

// Op makes the view's clip area report the pointer's moves. The area must
// contain the items: their own areas pass the moves on to it.
func (h *hoverPlay) Op(gtx layout.Context) {
	event.Op(gtx.Ops, h)
}

// Play reports whether the item of key, hovered or not, plays now; cheap is
// whether playing it starts no costly decoder.
func (h *hoverPlay) Play(gtx layout.Context, key any, hovered, cheap bool) bool {
	if !hovered {
		if h.key == key {
			h.key = nil
		}
		return false
	}
	if h.key != key {
		h.key, h.playing = key, false
		h.delay(gtx, hoverMoveDelay)
	}
	if h.scrolled() {
		h.playing = false
		h.delay(gtx, hoverScrollDelay)
	}
	if cheap {
		h.playing = true
	}
	if !h.playing {
		if gtx.Now.Before(h.due) {
			gtx.Execute(op.InvalidateCmd{At: h.due})
		} else {
			h.playing = true
		}
	}
	return h.playing
}

// delay puts off playing by at least d from now. A longer wait, as after a
// scroll, is kept.
func (h *hoverPlay) delay(gtx layout.Context, d time.Duration) {
	if due := gtx.Now.Add(d); due.After(h.due) {
		h.due = due
	}
}

// scrolled reports whether a list moved since the last call. A list lays
// out its items after taking its scroll, so this sees the frame's scroll.
func (h *hoverPlay) scrolled() bool {
	if len(h.offsets) != len(h.lists) {
		h.offsets = make([]layout.Position, len(h.lists))
		for i, l := range h.lists {
			h.offsets[i] = l.Position
		}
		return false
	}
	moved := false
	for i, l := range h.lists {
		was := h.offsets[i]
		if was.First != l.Position.First || was.Offset != l.Position.Offset {
			moved = true
		}
		h.offsets[i] = l.Position
	}
	return moved
}
