// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"os"
	"strings"
	"sync"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/layout"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

// reactionMenuStore is the menu store that also takes reactions, and keeps
// the ones toggled.
type reactionMenuStore struct {
	*menuStore
	mu        sync.Mutex
	available []model.Reaction
	toggled   []model.Reaction
}

func (s *reactionMenuStore) ChatReactions(int64) ([]model.Reaction, int, bool) {
	return s.available, 1, true
}
func (s *reactionMenuStore) ToggleReaction(msg model.Message, r model.Reaction, report func(error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toggled = append(s.toggled, r)
}

func newReactionHarness(t *testing.T, available int) (*menuHarness, *reactionMenuStore) {
	var store *reactionMenuStore
	h := newMenuHarnessOn(t, func(_ *menuStore, messages []model.Message) {
		messages[2].Reactions = []model.Reaction{{Emoji: "❤", Count: 3}, {Emoji: "👍", Count: 1, Chosen: true}}
	}, func(s *menuStore) model.ConversationStore {
		store = &reactionMenuStore{menuStore: s}
		for _, e := range demoEmoji[:available] {
			store.available = append(store.available, model.Reaction{Emoji: e})
		}
		return store
	})
	return h, store
}

var demoEmoji = strings.Fields("👍 ❤ 🔥 🥰 👏 😁 🤔 🤯 😱 🤬 😢 🎉 🤩 🤮 💩 🙏")

// TestReactionChipToggles clicks a reaction under a message.
func TestReactionChipToggles(t *testing.T) {
	h, store := newReactionHarness(t, 3)
	h.frame()
	r := h.page.rows[3]
	if r == nil || len(r.reactions) != 2 {
		t.Fatal("message 3 has no reaction chips")
	}
	// The chips are at the bottom of the bubble, the first at its start.
	at := h.messageAt(3)
	found := false
	for y := at.Y; y < at.Y+60 && !found; y += 2 {
		for x := float32(20); x < 200; x += 4 {
			h.press(pointer.ButtonPrimary, f32.Pt(x, y))
			store.mu.Lock()
			found = len(store.toggled) > 0
			store.mu.Unlock()
			if found {
				break
			}
		}
	}
	if !found {
		t.Fatal("no click on the message toggled a reaction")
	}
	if got := store.toggled[0]; got.Emoji != "❤" {
		t.Errorf("toggled %+v, want ❤ first", got)
	}
}

// TestMenuReactionStrip picks a reaction from the message menu, first from
// the row and then from the expanded list.
func TestMenuReactionStrip(t *testing.T) {
	h, store := newReactionHarness(t, len(demoEmoji))
	h.openMenu(5)
	m := &h.page.messageMenu
	if len(m.reactions.shown) != len(demoEmoji) || m.reactions.collapsedCount() != reactionsPerRow-1 {
		t.Fatalf("strip shows %d, %d in its row", len(m.reactions.shown), m.reactions.collapsedCount())
	}
	cellAt := func(i int) f32.Point {
		x0 := max(reactionPadding, (m.rect.Dx()-reactionsPerRow*reactionCell)/2)
		row, col := i/reactionsPerRow, i%reactionsPerRow
		return f32.Pt(float32(m.rect.Min.X+x0+col*reactionCell+reactionCell/2), float32(m.rect.Min.Y+reactionPadding+row*reactionCell+reactionCell/2))
	}
	h.press(pointer.ButtonPrimary, cellAt(2))
	if len(store.toggled) != 1 || store.toggled[0].Emoji != demoEmoji[2] {
		t.Fatalf("toggled %+v", store.toggled)
	}
	if m.open {
		t.Error("the menu stayed open")
	}
	h.openMenu(5)
	h.press(pointer.ButtonPrimary, cellAt(reactionsPerRow-1))
	h.frames(2)
	if !m.open || !m.reactions.expanded {
		t.Fatal("the expand button did not expand the strip")
	}
	h.press(pointer.ButtonPrimary, cellAt(reactionsPerRow+1))
	if len(store.toggled) != 2 || store.toggled[1].Emoji != demoEmoji[reactionsPerRow+1] {
		t.Fatalf("toggled %+v", store.toggled)
	}
}

// TestRenderReactionMenu saves the menu with its reactions, collapsed and
// expanded: MENU_PNG=/tmp/menu.png writes menu-reactions*.png beside it.
func TestRenderReactionMenu(t *testing.T) {
	path := os.Getenv("MENU_PNG")
	if path == "" {
		t.Skip("set MENU_PNG to a file")
	}
	h, _ := newReactionHarness(t, len(demoEmoji))
	p := h.page
	l := localization.For("ru")
	for _, expanded := range []bool{false, true} {
		m := &p.messageMenu
		m.open, m.id, m.at, m.top = true, 3, image.Pt(140, 200), 34
		m.reactions.expanded = expanded
		file := strings.TrimSuffix(path, ".png") + "-reactions.png"
		if expanded {
			file = strings.TrimSuffix(path, ".png") + "-reactions-expanded.png"
		}
		renderFrames(t, image.Pt(500, 700), file, func(gtx layout.Context) {
			p.images.BeginFrame()
			p.media.BeginFrame()
			p.Layout(gtx, model.Chat{ID: 1, Title: "Чат", Kind: model.KindGroup}, l, false)
			p.layoutDialogs(gtx, l)
			p.media.EndFrame()
			p.images.EndFrame()
		})
	}
}

// lateReactionStore has no reactions for its chat until loaded is set, as a
// store whose reactions are still loading.
type lateReactionStore struct {
	reactionMenuStore
	loaded bool
}

func (s *lateReactionStore) ChatReactions(chat int64) ([]model.Reaction, int, bool) {
	if !s.loaded {
		return nil, 0, false
	}
	return s.reactionMenuStore.ChatReactions(chat)
}

// TestMenuReactionsArriveOpen checks that the reactions may load while the
// menu is open: the strip shows them in the frame they come, with nothing
// updated before it.
func TestMenuReactionsArriveOpen(t *testing.T) {
	var store *lateReactionStore
	h := newMenuHarnessOn(t, nil, func(s *menuStore) model.ConversationStore {
		store = &lateReactionStore{reactionMenuStore: reactionMenuStore{menuStore: s}}
		for _, e := range demoEmoji {
			store.available = append(store.available, model.Reaction{Emoji: e})
		}
		return store
	})
	h.openMenu(5)
	if len(h.page.messageMenu.reactions.shown) != 0 {
		t.Fatal("reactions shown before they loaded")
	}
	store.loaded = true
	h.frames(3)
	if got := len(h.page.messageMenu.reactions.shown); got != len(demoEmoji) {
		t.Fatalf("strip shows %d reactions", got)
	}
}
