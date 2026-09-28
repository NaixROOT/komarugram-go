// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

// A deletion Telegram tells of leaves the message in the open history,
// marked deleted, when deleted messages are kept; a chat with a bot keeps
// none; the account's own deletion deletes.
func TestKeepDeletedFromUpdates(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	s.rememberPeers([]tg.UserClass{&tg.User{ID: 5, FirstName: "Ann"}, &tg.User{ID: 6, Bot: true, FirstName: "Bot"}}, nil)
	s.SetKeep(model.Keep{Deleted: true, Edits: true})
	for _, chat := range []int64{5, 6} {
		s.history.histories[chat] = &model.History{}
	}
	raw := []tg.MessageClass{
		&tg.Message{ID: 10, PeerID: &tg.PeerUser{UserID: 5}, Message: "keep me", Date: 10},
		&tg.Message{ID: 11, PeerID: &tg.PeerUser{UserID: 5}, Message: "after", Date: 11},
		&tg.Message{ID: 12, PeerID: &tg.PeerUser{UserID: 6}, Message: "bot", Date: 12},
	}
	msgs, err := s.ingest(ctx, raw, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range msgs {
		h := s.history.histories[m.Key.ChatID]
		h.Messages = append(h.Messages, m)
	}
	if err := s.Handle(ctx, &tg.UpdateShort{Update: &tg.UpdateDeleteMessages{Messages: []int{10, 12}}}); err != nil {
		t.Fatal(err)
	}
	got := s.History(5).Messages
	if len(got) != 2 || !got[0].Deleted || got[0].Text != "keep me" || got[1].Deleted {
		t.Fatalf("the history is %+v", got)
	}
	if got := s.History(6).Messages; len(got) != 0 {
		t.Fatalf("a bot's deleted message stayed: %+v", got)
	}
	// A late page of Telegram's does not bring the old version back.
	late, err := s.ingest(ctx, raw[:1], false, 0)
	if err != nil || len(late) != 0 {
		t.Fatalf("a late page gave %+v, %v", late, err)
	}
	if m, ok, _ := s.history.cache.Message(ctx, 5, 10); !ok || !m.Deleted {
		t.Fatalf("the cache has %+v", m)
	}
}
