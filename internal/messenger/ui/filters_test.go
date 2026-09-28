// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"testing"

	"komarugram/internal/messenger/model"
	"komarugram/internal/messenger/preferences"
)

func TestMessageFilter(t *testing.T) {
	f := compileFilter(preferences.Filters{Enabled: true, HideBlocked: true, Patterns: []preferences.FilterPattern{
		{Text: "promo", CaseInsensitive: true},
		{Text: "^#news", Reversed: true, Chat: 7},
		{Text: "(broken"},
	}})
	blocked := func(id int64) bool { return id == 9 }
	msg := func(text string) model.Message { return model.Message{Text: text, SenderID: 3} }
	for _, c := range []struct {
		name string
		m    model.Message
		chat int64
		kind model.ChatKind
		want bool
	}{
		{"a pattern in a channel", msg("Big PROMO today"), 1, model.KindChannel, true},
		{"no match", msg("hello"), 1, model.KindChannel, false},
		{"not in groups without InChats", msg("promo"), 2, model.KindGroup, false},
		{"reversed, in its chat", msg("hello"), 7, model.KindChannel, true},
		{"reversed, matching", msg("#news today"), 7, model.KindChannel, false},
		{"reversed, elsewhere", msg("hello"), 8, model.KindChannel, false},
		{"own messages stay", model.Message{Text: "promo", Outgoing: true}, 1, model.KindChannel, false},
		{"a blocked sender, in a group", model.Message{Text: "hi", SenderID: 9}, 2, model.KindGroup, true},
		{"a forward from a blocked user", model.Message{Text: "hi", SenderID: 3, ForwardFromID: 9}, 2, model.KindGroup, true},
		{"the private chat with a blocked user", model.Message{Text: "hi", SenderID: 9}, 9, model.KindUser, false},
		{"a poll's question", model.Message{Poll: &model.Poll{Question: "promo?"}}, 1, model.KindChannel, true},
	} {
		if got := f.hides(c.m, c.chat, c.kind, blocked); got != c.want {
			t.Errorf("%s: hides %v, want %v", c.name, got, c.want)
		}
	}
	f.source.InChats = true
	if !f.hides(msg("promo"), 2, model.KindGroup, blocked) {
		t.Error("InChats does not filter groups")
	}
	f.source.Enabled = false
	if f.hides(msg("promo"), 1, model.KindChannel, blocked) {
		t.Error("disabled filters hide")
	}
}

// The history leaves out what the filter hides, and the chat's menu shows
// it again.
func TestHistoryFiltered(t *testing.T) {
	h := newMenuHarness(t, func(_ *menuStore, messages []model.Message) {
		messages[4].Text = "buy now, promo"
		messages[6].Text = "PROMO again"
	})
	p := h.page
	p.filter = compileFilter(preferences.Filters{Enabled: true, InChats: true, Patterns: []preferences.FilterPattern{{Text: "promo", CaseInsensitive: true}}})
	h.frames(2)
	if len(p.messages) != 18 || p.filtered != 2 {
		t.Fatalf("%d messages shown, %d filtered", len(p.messages), p.filtered)
	}
	actions := p.chatMenuActions()
	found := false
	for _, a := range actions {
		found = found || a == chatMenuFiltered
	}
	if !found {
		t.Fatal("the chat's menu does not offer the filtered messages")
	}
	p.showFiltered = map[int64]bool{1: true}
	p.revision = 0
	h.frames(2)
	if len(p.messages) != 20 {
		t.Fatalf("%d messages shown with the filtered ones", len(p.messages))
	}
}
