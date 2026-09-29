// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"komarugram/internal/messenger/model"
)

func convertOne(t *testing.T, m *tg.Message) model.Message {
	t.Helper()
	s := testStore(t)
	msgs, err := s.ingest(context.Background(), []tg.MessageClass{m}, false, 0)
	if err != nil || len(msgs) != 1 {
		t.Fatalf("ingest: %v, %d messages", err, len(msgs))
	}
	return msgs[0]
}

// The buttons under a message keep what pressing them needs.
func TestInlineKeyboardConverted(t *testing.T) {
	callback := &tg.KeyboardButtonCallback{Text: "Refresh", Data: []byte{1, 2, 3}}
	secret := &tg.KeyboardButtonCallback{Text: "Pay", Data: []byte{9}, RequiresPassword: true}
	m := convertOne(t, &tg.Message{ID: 7, PeerID: &tg.PeerUser{UserID: 5}, Message: "menu", Date: 10, ReplyMarkup: &tg.ReplyInlineMarkup{Rows: []tg.KeyboardButtonRow{
		{Buttons: []tg.KeyboardButtonClass{&tg.KeyboardButtonURL{Text: "Site", URL: "https://example.org"}, callback}},
		{Buttons: []tg.KeyboardButtonClass{&tg.KeyboardButtonCopy{Text: "Code", CopyText: "123456"}, secret, &tg.KeyboardButtonBuy{Text: "Buy"}}},
	}}})
	want := [][]model.MessageButton{
		{{Text: "Site", Kind: "url", URL: "https://example.org"}, {Text: "Refresh", Kind: "callback", Data: []byte{1, 2, 3}}},
		{{Text: "Code", Kind: "copy", Copy: "123456"}, {Text: "Pay", Kind: "action"}, {Text: "Buy", Kind: "action"}},
	}
	if len(m.Buttons) != 2 || len(m.Buttons[0]) != 2 || len(m.Buttons[1]) != 3 {
		t.Fatalf("buttons %+v", m.Buttons)
	}
	for y := range want {
		for x, w := range want[y] {
			g := m.Buttons[y][x]
			if g.Text != w.Text || g.Kind != w.Kind || g.URL != w.URL || g.Copy != w.Copy || string(g.Data) != string(w.Data) {
				t.Errorf("button %d,%d is %+v, want %+v", y, x, g, w)
			}
		}
	}
	if m.Keyboard != nil {
		t.Fatalf("an inline keyboard came as a reply keyboard: %+v", m.Keyboard)
	}
}

// A reply keyboard is the chat's, not the message's buttons; another
// message takes it away.
func TestReplyKeyboardConverted(t *testing.T) {
	m := convertOne(t, &tg.Message{ID: 8, PeerID: &tg.PeerUser{UserID: 5}, Message: "pick", Date: 10, ReplyMarkup: &tg.ReplyKeyboardMarkup{
		SingleUse: true, Placeholder: "Choose",
		Rows: []tg.KeyboardButtonRow{{Buttons: []tg.KeyboardButtonClass{&tg.KeyboardButton{Text: "Yes"}, &tg.KeyboardButtonRequestPhone{Text: "Share"}}}},
	}})
	if len(m.Buttons) != 0 || m.Keyboard == nil || !m.Keyboard.SingleUse || m.Keyboard.Placeholder != "Choose" {
		t.Fatalf("message %+v", m)
	}
	row := m.Keyboard.Rows[0]
	if row[0].Kind != "text" || row[0].Text != "Yes" || row[1].Kind != "action" {
		t.Fatalf("row %+v", row)
	}
	gone := convertOne(t, &tg.Message{ID: 9, PeerID: &tg.PeerUser{UserID: 5}, Message: "bye", Date: 11, ReplyMarkup: &tg.ReplyKeyboardHide{}})
	if !gone.KeyboardHide || gone.Keyboard != nil {
		t.Fatalf("message %+v", gone)
	}
}

// Pressing a button asks the bot for the message it is under, with its data,
// and passes on the answer; a bot that stays silent is told as such.
func TestPressButton(t *testing.T) {
	s := testStore(t)
	s.history.peers[5] = peerRecord{Kind: "user", ID: 5, Hash: 55}
	var asked *tg.MessagesGetBotCallbackAnswerRequest
	var fail error
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		r, ok := in.(*tg.MessagesGetBotCallbackAnswerRequest)
		if !ok {
			return nil, errors.New("unexpected request")
		}
		asked = r
		if fail != nil {
			return nil, fail
		}
		answer := &tg.MessagesBotCallbackAnswer{Alert: true, HasURL: true}
		answer.SetMessage("Done")
		answer.SetURL("https://example.org/next")
		return answer, nil
	})
	key := model.MessageKey{ChatID: 5, MessageID: 42}
	got, err := s.PressButton(context.Background(), key, []byte{7})
	if err != nil {
		t.Fatal(err)
	}
	if asked.MsgID != 42 || string(asked.Data) != "\x07" {
		t.Fatalf("asked %+v", asked)
	}
	if user, ok := asked.Peer.(*tg.InputPeerUser); !ok || user.UserID != 5 || user.AccessHash != 55 {
		t.Fatalf("asked of %+v", asked.Peer)
	}
	if got.Text != "Done" || !got.Alert || got.URL != "https://example.org/next" {
		t.Fatalf("answer %+v", got)
	}
	fail = tgerr.New(400, "BOT_RESPONSE_TIMEOUT")
	if _, err := s.PressButton(context.Background(), key, nil); !errors.Is(err, model.ErrBotSilent) {
		t.Fatalf("silent bot: %v", err)
	}
}

// The chat's keyboard is the newest one set; hiding, or a single-use one that
// was answered, takes it away.
func TestActiveKeyboard(t *testing.T) {
	kb := &model.ReplyKeyboard{Rows: [][]model.MessageButton{{{Text: "a", Kind: "text"}}}}
	once := &model.ReplyKeyboard{Rows: kb.Rows, SingleUse: true}
	msg := func(k *model.ReplyKeyboard, hide, out bool) model.Message {
		return model.Message{Keyboard: k, KeyboardHide: hide, Outgoing: out}
	}
	for name, tc := range map[string]struct {
		messages []model.Message
		want     *model.ReplyKeyboard
	}{
		"none":            {[]model.Message{msg(nil, false, false)}, nil},
		"set":             {[]model.Message{msg(kb, false, false), msg(nil, false, true)}, kb},
		"newest wins":     {[]model.Message{msg(once, false, false), msg(kb, false, false)}, kb},
		"hidden":          {[]model.Message{msg(kb, false, false), msg(nil, true, false)}, nil},
		"set after hide":  {[]model.Message{msg(nil, true, false), msg(kb, false, false)}, kb},
		"once unanswered": {[]model.Message{msg(once, false, false), msg(nil, false, false)}, once},
		"once answered":   {[]model.Message{msg(once, false, false), msg(nil, false, true)}, nil},
	} {
		if got, _ := model.ActiveKeyboard(tc.messages); got != tc.want {
			t.Errorf("%s: got %+v, want %+v", name, got, tc.want)
		}
	}
}

// The commands of a bot come from its full user, on the first ask, and are
// kept.
func TestBotCommands(t *testing.T) {
	s := testStore(t)
	s.history.peers[5] = peerRecord{Kind: "user", ID: 5, Hash: 55}
	var mu sync.Mutex
	asks := 0
	changed := make(chan struct{}, 4)
	s.changed = func() { changed <- struct{}{} }
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		r, ok := in.(*tg.UsersGetFullUserRequest)
		if !ok {
			return nil, errors.New("unexpected request")
		}
		if user, ok := r.ID.(*tg.InputUser); !ok || user.UserID != 5 || user.AccessHash != 55 {
			return nil, errors.New("asked of another user")
		}
		mu.Lock()
		asks++
		mu.Unlock()
		info := tg.BotInfo{}
		info.SetCommands([]tg.BotCommand{{Command: "start", Description: "Begin"}, {Command: "help", Description: "What it does"}})
		full := &tg.UsersUserFull{}
		full.FullUser.SetBotInfo(info)
		return full, nil
	})
	if got := s.BotCommands(5); got != nil {
		t.Fatalf("commands %v before they were read", got)
	}
	select {
	case <-changed:
	case <-time.After(3 * time.Second):
		t.Fatal("the window was not told of the commands")
	}
	got := s.BotCommands(5)
	if len(got) != 2 || got[0] != (model.BotCommand{Command: "start", Description: "Begin"}) {
		t.Fatalf("commands %v", got)
	}
	s.BotCommands(5)
	mu.Lock()
	defer mu.Unlock()
	if asks != 1 {
		t.Fatalf("asked %d times", asks)
	}
}

func TestMatchCommands(t *testing.T) {
	cmds := []model.BotCommand{{Command: "start"}, {Command: "Settings"}, {Command: "stop"}}
	names := func(list []model.BotCommand) (out []string) {
		for _, c := range list {
			out = append(out, c.Command)
		}
		return out
	}
	for text, want := range map[string]string{
		"/":       "start Settings stop",
		"/st":     "start stop",
		"/SET":    "Settings",
		"/x":      "",
		"hello":   "",
		"/start ": "",
		"":        "",
	} {
		if got := strings.Join(names(model.MatchCommands(cmds, text)), " "); got != want {
			t.Errorf("%q matched %q, want %q", text, got, want)
		}
	}
}
