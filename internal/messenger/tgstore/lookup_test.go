// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

// waitLookup asks for a message until the lookup ends.
func waitLookup(t *testing.T, s *Store, chat int64, id model.MessageID) (model.Message, model.LookupState) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		m, state := s.LookupMessage(chat, id)
		if state != model.LookupLoading {
			return m, state
		}
		if time.Now().After(deadline) {
			t.Fatalf("message %d of %d still loading", id, chat)
		}
		time.Sleep(time.Millisecond)
	}
}

// A message the history has not loaded is looked up from Telegram once,
// kept apart from the history, and gone when Telegram deletes it.
func TestLookupMessage(t *testing.T) {
	s := testStore(t)
	const channel = 77
	chat := peerID(&tg.PeerChannel{ChannelID: channel})
	s.history.peers[chat] = peerRecord{Kind: "channel", ID: channel, Hash: 7}
	asked := 0
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		r, ok := in.(*tg.ChannelsGetMessagesRequest)
		if !ok {
			return nil, fmt.Errorf("unexpected %T", in)
		}
		asked++
		id := r.ID[0].(*tg.InputMessageID).ID
		if id == 3 {
			return &tg.MessagesChannelMessages{Messages: []tg.MessageClass{&tg.MessageEmpty{ID: 3}}}, nil
		}
		return &tg.MessagesChannelMessages{Messages: []tg.MessageClass{&tg.Message{ID: id, PeerID: &tg.PeerChannel{ChannelID: channel}, Message: "pinned long ago", Date: 10}}}, nil
	})
	m, state := waitLookup(t, s, chat, 2)
	if state != model.LookupFound || m.Text != "pinned long ago" || m.Key.MessageID != 2 {
		t.Fatalf("found %v %+v", state, m)
	}
	if _, state = s.LookupMessage(chat, 2); state != model.LookupFound || asked != 1 {
		t.Fatalf("a second ask is %v after %d requests", state, asked)
	}
	if _, ok, _ := s.history.cache.Message(context.Background(), chat, 2); ok {
		t.Fatal("a message looked up joined the history's cache")
	}
	if _, state = waitLookup(t, s, chat, 3); state != model.LookupGone {
		t.Fatalf("a deleted message is %v", state)
	}
	if _, err := s.deleteMessages(context.Background(), chat, []int{2}); err != nil {
		t.Fatal(err)
	}
	if _, state = s.LookupMessage(chat, 2); state != model.LookupGone {
		t.Fatalf("a message deleted after its lookup is %v", state)
	}
}

// A message in the history's cache needs no request.
func TestLookupMessageFromCache(t *testing.T) {
	s := testStore(t)
	chat := int64(5)
	if _, err := s.ingest(context.Background(), []tg.MessageClass{&tg.Message{ID: 4, PeerID: &tg.PeerUser{UserID: 5}, Message: "cached", Date: 10}}, false, 0); err != nil {
		t.Fatal(err)
	}
	if m, state := waitLookup(t, s, chat, 4); state != model.LookupFound || m.Text != "cached" {
		t.Fatalf("found %v %+v", state, m)
	}
}
