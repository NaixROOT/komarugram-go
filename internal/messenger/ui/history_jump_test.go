// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"testing"
	"time"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"
)

var errFailed = errors.New("failed")

// jumpSource is a history that can be opened at its ends, and reloaded,
// and tells which it was asked for.
type jumpSource struct {
	benchmarkHistory
	calls *[]string
	// loading, if set, is what a jump starts: the history loads while it is.
	loading *bool
}

func (s jumpSource) HistorySince(chat int64, revision uint64) (model.History, bool) {
	h, fresh := s.benchmarkHistory.HistorySince(chat, revision)
	if s.loading != nil && *s.loading {
		h.LoadingOlder = true
	}
	return h, fresh
}

func (s jumpSource) RevealFirst(int64) bool { return s.reveal("first") }
func (s jumpSource) RevealLast(int64) bool  { return s.reveal("last") }

func (s jumpSource) reveal(call string) bool {
	*s.calls = append(*s.calls, call)
	if s.loading != nil {
		*s.loading = true
	}
	return true
}
func (s jumpSource) Reload(int64) { *s.calls = append(*s.calls, "reload") }

// jumpButton is the centre of the n-th button from the bottom in the
// harness's window: 40 dp round, 16 dp from the corner of the history above
// its composer, 12 dp apart.
func (h *chatInputHarness) jumpButton(n int) f32.Point {
	bottom := 600
	if c := h.page.composer; c != nil {
		bottom = c.top
	}
	return f32.Pt(100+400-16-20, float32(bottom-16-20-n*(40+12)))
}

func (h *chatInputHarness) click(at f32.Point) {
	h.send(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: at, Buttons: pointer.ButtonPrimary})
	h.send(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: at})
	h.frame()
}

func TestJumpButtonsScroll(t *testing.T) {
	// At the start of 100 messages, none older: more than one request's.
	h := newChatInputHarnessOf(t, 100, func(m model.History) model.ConversationStore { return benchmarkHistory{h: m} })
	p := h.page
	if !p.atStart(model.History{}) || p.atEnd(model.History{}) {
		t.Fatalf("at the start: atStart %v, atEnd %v", p.atStart(model.History{}), p.atEnd(model.History{}))
	}
	h.click(h.jumpButton(0))
	if p.list.Position.BeforeEnd {
		t.Fatal("the button to the end left the history before it")
	}
	if !p.atEnd(model.History{}) || p.atStart(model.History{}) {
		t.Fatalf("at the end: atStart %v, atEnd %v", p.atStart(model.History{}), p.atEnd(model.History{}))
	}
	// Only one button is left, the one to the start, and it is the lowest.
	h.click(h.jumpButton(0))
	if p.list.Position.First != 0 || p.list.Position.Offset != 0 {
		t.Fatalf("the button to the start stopped at %d+%d", p.list.Position.First, p.list.Position.Offset)
	}
}

func TestJumpButtonsHideWhenThereIsNowhereToGo(t *testing.T) {
	h := newChatInputHarnessOf(t, 2, func(m model.History) model.ConversationStore { return benchmarkHistory{h: m} })
	h.page.list.Position.BeforeEnd = false
	h.frame()
	if !h.page.atStart(model.History{}) || !h.page.atEnd(model.History{}) {
		t.Fatalf("two messages that fit: atStart %v, atEnd %v", h.page.atStart(model.History{}), h.page.atEnd(model.History{}))
	}
}

func TestJumpButtonsReachUnloadedEnds(t *testing.T) {
	for n, want := range []string{"last", "first"} {
		var calls []string
		source := func(m model.History) model.ConversationStore {
			m.HasOlder, m.HasNewer = true, true
			return jumpSource{benchmarkHistory{h: m}, &calls, nil}
		}
		h := newChatInputHarnessOf(t, 40, source)
		// Both ends are unloaded, so both buttons are there: the end below.
		h.click(h.jumpButton(n))
		if len(calls) != 1 || calls[0] != want {
			t.Fatalf("button %d: the store was asked for %v, want %s", n, calls, want)
		}
	}
}

func TestRetryButton(t *testing.T) {
	var calls []string
	source := func(m model.History) model.ConversationStore {
		m.Err = errFailed
		return jumpSource{benchmarkHistory{h: m}, &calls, nil}
	}
	h := newChatInputHarnessOf(t, 40, source)
	// At the start with a failure: the end, then the retry above it.
	h.click(h.jumpButton(1))
	if len(calls) != 1 || calls[0] != "reload" {
		t.Fatalf("the store was asked for %v, want a reload", calls)
	}
}

func (s jumpSource) LoadOlder(int64) {}
func (s jumpSource) LoadNewer(int64) {}

func TestNoButtonToTheStartOfAChatOfOneRequest(t *testing.T) {
	h := newChatInputHarness(t) // 40 messages, all there is
	p := h.page
	h.click(h.jumpButton(0)) // the one to the end
	if p.list.Position.BeforeEnd {
		t.Fatal("the button to the end did nothing")
	}
	if p.atStart(model.History{}) || !p.oneRequest(model.History{}) {
		t.Fatalf("not at the start of a chat of one request: atStart %v, oneRequest %v", p.atStart(model.History{}), p.oneRequest(model.History{}))
	}
	first, offset := p.list.Position.First, p.list.Position.Offset
	h.click(h.jumpButton(0)) // nothing is there now
	if p.list.Position.First != first || p.list.Position.Offset != offset {
		t.Fatal("a button to the start is drawn in a chat of one request")
	}
}

// TestRenderJumpButtons saves the buttons over the history, for looking at
// them: JUMP_PNG_DIR=/tmp/jump. In the middle of a chat with older and newer
// messages, with a load on its way and with one failed.
func TestRenderJumpButtons(t *testing.T) {
	dir := os.Getenv("JUMP_PNG_DIR")
	if dir == "" {
		t.Skip("set JUMP_PNG_DIR to a directory")
	}
	for name, set := range map[string]func(*model.History){
		"middle":  func(h *model.History) {},
		"loading": func(h *model.History) { h.LoadingOlder = true },
		"failed":  func(h *model.History) { h.Err = errFailed },
	} {
		var calls []string
		var messages []model.Message
		for i := 1; i <= 100; i++ {
			messages = append(messages, model.Message{Key: model.MessageKey{ChatID: 1, MessageID: model.MessageID(i)}, Date: time.Unix(1700000000, 0), Text: fmt.Sprintf("Message %d", i), ContentRevision: 1})
		}
		hist := model.History{Messages: messages, Revision: 1, HasOlder: true, HasNewer: true}
		set(&hist)
		p := newChatPage(jumpSource{benchmarkHistory{h: hist}, &calls, nil}, func() {})
		p.list.Position.First, p.list.Position.BeforeEnd = 50, true
		renderFrames(t, image.Pt(500, 800), filepath.Join(dir, "jump-"+name+".png"), func(gtx layout.Context) {
			p.Layout(gtx, model.Chat{ID: 1, Title: "Чат", Kind: model.KindGroup}, localization.For("ru"), false)
		})
		p.Close()
	}
}

// A thread's history is not opened again by the store from a viewport, so the
// page puts it at the end asked for once the load of the jump is over.
func TestJumpInAThread(t *testing.T) {
	for n, want := range []struct {
		call string
		to   jump
	}{{"last", jumpEnd}, {"first", jumpStart}} {
		var calls []string
		loading := new(bool)
		source := func(m model.History) model.ConversationStore {
			m.HasOlder, m.HasNewer, m.ThreadRoot = true, true, 7
			return jumpSource{benchmarkHistory{h: m}, &calls, loading}
		}
		h := newChatInputHarnessOf(t, 100, source)
		p := h.page
		h.click(h.jumpButton(n))
		if len(calls) != 1 || calls[0] != want.call {
			t.Fatalf("button %d: the store was asked for %v", n, calls)
		}
		// Loading: the jump waits, and the thread is not forgotten, for the
		// store would open it at the end again.
		if p.jump != want.to || p.chat != 1 {
			t.Fatalf("button %d: while loading, jump %d, chat %d", n, p.jump, p.chat)
		}
		p.list.Position.First, p.list.Position.BeforeEnd = 30, true
		*loading = false
		h.frame()
		if p.jump != jumpNone {
			t.Fatalf("button %d: the jump waits still", n)
		}
		if want.to == jumpStart && (p.list.Position.First != 0 || p.list.Position.Offset != 0) || want.to == jumpEnd && p.list.Position.BeforeEnd {
			t.Fatalf("button %d: the history is at %d+%d, before the end %v", n, p.list.Position.First, p.list.Position.Offset, p.list.Position.BeforeEnd)
		}
	}
}
