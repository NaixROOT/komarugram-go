// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"

	"komarugram/internal/messenger/mockstore"
	"komarugram/internal/messenger/model"
)

// botStore is a menu store whose bot answers what it is told to.
type botStore struct {
	*menuStore
	mu      sync.Mutex
	pressed []model.MessageKey
	data    [][]byte
	answer  model.BotAnswer
	err     error
}

func (s *botStore) PressButton(_ context.Context, key model.MessageKey, data []byte) (model.BotAnswer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pressed, s.data = append(s.pressed, key), append(s.data, data)
	return s.answer, s.err
}

func (s *botStore) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.pressed)
}

// pressButtonOf clicks down the height of message id, at x, until a button
// was pressed or nothing is left to hit.
func (h *menuHarness) pressButtonOf(id model.MessageID, x float32, store *botStore) {
	h.t.Helper()
	at := h.messageAt(id)
	for y := at.Y - 80; y < at.Y+120 && store.count() == 0; y += 4 {
		for x := x; x < 300 && store.count() == 0; x += 60 {
			h.press(pointer.ButtonPrimary, f32.Pt(x, y))
		}
	}
}

func botHarness(t *testing.T, answer model.BotAnswer, err error, buttons [][]model.MessageButton) (*menuHarness, *botStore) {
	var store *botStore
	h := newMenuHarnessOn(t, func(_ *menuStore, messages []model.Message) {
		messages[2].Buttons = buttons
		messages[2].SenderName = "Bot"
	}, func(s *menuStore) model.ConversationStore {
		store = &botStore{menuStore: s, answer: answer, err: err}
		return store
	})
	h.frames(3)
	return h, store
}

func waitFor(t *testing.T, h *menuHarness, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for !ok() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out: %s", what)
		}
		h.frames(1)
		time.Sleep(time.Millisecond)
	}
}

// A callback button asks the bot for the message it is under, with its data,
// and the bot's notice comes to the toast; a link it answers is offered.
func TestCallbackButton(t *testing.T) {
	buttons := [][]model.MessageButton{{{Text: "Refresh", Kind: "callback", Data: []byte{7, 8}}}}
	h, store := botHarness(t, model.BotAnswer{Text: "Refreshed"}, nil, buttons)
	h.pressButtonOf(3, 60, store)
	if store.count() != 1 {
		t.Fatal("the button did not ask the bot")
	}
	if store.pressed[0].MessageID != 3 || string(store.data[0]) != "\x07\x08" {
		t.Fatalf("asked %v with %v", store.pressed, store.data)
	}
	waitFor(t, h, "the bot's notice in the toast", func() bool { return h.page.toast.Text() == "Refreshed" })

	h, store = botHarness(t, model.BotAnswer{URL: "https://example.org/x"}, nil, buttons)
	h.pressButtonOf(3, 60, store)
	waitFor(t, h, "the link the bot answered", func() bool { return h.page.link == "https://example.org/x" })
}

// A bot that stays silent, or a failure, is told in the toast.
func TestCallbackButtonFailures(t *testing.T) {
	buttons := [][]model.MessageButton{{{Text: "Refresh", Kind: "callback", Data: []byte{1}}}}
	h, store := botHarness(t, model.BotAnswer{}, model.ErrBotSilent, buttons)
	h.pressButtonOf(3, 60, store)
	waitFor(t, h, "the silent bot in the toast", func() bool { return h.page.toast.Text() == "The bot did not answer" })
}

// A button that is not the client's to press does nothing.
func TestActionButtonIsDisabled(t *testing.T) {
	buttons := [][]model.MessageButton{{{Text: "Pay", Kind: "action"}}}
	h, store := botHarness(t, model.BotAnswer{}, nil, buttons)
	h.pressButtonOf(3, 60, store)
	h.frames(10)
	if store.count() != 0 {
		t.Fatal("a button the client does not do asked the bot")
	}
}

// waitBotPage draws the page of the demo's notes bot until ok.
func waitBotPage(t *testing.T, h *composerHarness, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for !ok() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out: %s", what)
		}
		h.frame()
		time.Sleep(time.Millisecond)
	}
}

// An empty chat with a bot has Start in place of the composer; the bot's
// answer to /start sets a keyboard under it, whose keys send their text, and
// a message that hides it takes it away.
func TestBotStartAndKeyboard(t *testing.T) {
	h := newComposerHarness(t)
	h.chat, h.kind = mockstore.DemoNotesBot, model.KindBot
	h.frame()
	waitBotPage(t, h, "the empty chat of a bot", func() bool { return h.p.bot.empty })
	// Start takes the composer's bar.
	h.click(340, 680)
	waitBotPage(t, h, "the keyboard the bot sets", func() bool { return h.p.bot.keyboard != nil })
	if h.p.bot.empty {
		t.Fatal("the chat is still empty after /start")
	}
	rows := h.p.bot.keyboard.Rows
	if len(rows) != 2 || rows[1][0].Text != "Настройки" {
		t.Fatalf("keyboard %+v", rows)
	}
	// The key that says Settings is in the second row, first column.
	// The key that says Settings is in the second row: find it from the top
	// of the page, as the keys before it send their text and change nothing.
	for y := float32(440); y < 700 && h.p.bot.keyboard != nil; y += 10 {
		for range 5 {
			h.frame() // What was sent is done sending.
		}
		h.click(100, y)
	}
	waitBotPage(t, h, "the keyboard taken away", func() bool { return h.p.bot.keyboard == nil })
	if h.p.bot.keyboard != nil {
		t.Fatal("the keyboard is still there")
	}
}

// Typing "/" in a bot's chat lists its commands the text starts; a click on
// one sends it, and the field is left empty.
func TestBotCommandsMenu(t *testing.T) {
	h := newComposerHarness(t)
	h.chat, h.kind = 6, model.KindBot
	h.frame()
	h.click(160, 680)
	h.router.Queue(key.EditEvent{Text: "/st"})
	h.frame()
	h.frame()
	sent := func() []string {
		var texts []string
		for _, m := range h.p.source.History(6).Messages {
			if m.Outgoing && strings.HasPrefix(m.Text, "/") {
				texts = append(texts, m.Text)
			}
		}
		return texts
	}
	if len(sent()) != 0 {
		t.Fatalf("sent %v before a command was chosen", sent())
	}
	// The menu is over the composer, whose top is the bottom of its rows.
	top := float32(h.p.composer.top)
	for y := top - 40; y > top-140 && len(sent()) == 0; y -= 6 {
		h.click(100, y)
	}
	got := sent()
	if len(got) != 1 || (got[0] != "/start" && got[0] != "/stop") {
		t.Fatalf("sent %v, want one of the commands that begin /st", got)
	}
	h.frame()
	if text := h.p.composer.draft(6).editor.Text(); text != "" {
		t.Fatalf("the field still says %q", text)
	}
}
