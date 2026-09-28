// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

// waitPinned asks for chat's pinned messages until they are n.
func waitPinned(t *testing.T, s *Store, chat int64, n int) []model.MessageID {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		ids := s.PinnedMessages(chat)
		if len(ids) == n {
			return ids
		}
		if time.Now().After(deadline) {
			t.Fatalf("pinned messages are %v, want %d", ids, n)
		}
		time.Sleep(time.Millisecond)
	}
}

// Pinned messages load with the pinned filter, follow updates, stay hidden
// until a newer pin, and the bar's messages need no request of their own.
func TestPinnedMessages(t *testing.T) {
	s := testStore(t)
	chat := int64(5)
	s.history.peers[chat] = peerRecord{Kind: "user", ID: 5, Hash: 55}
	searched := 0
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		r, ok := in.(*tg.MessagesSearchRequest)
		if !ok {
			return nil, fmt.Errorf("unexpected %T", in)
		}
		if _, ok := r.Filter.(*tg.InputMessagesFilterPinned); !ok {
			return nil, fmt.Errorf("filter %T", r.Filter)
		}
		searched++
		msg := func(id int) *tg.Message {
			return &tg.Message{ID: id, PeerID: &tg.PeerUser{UserID: 5}, Message: fmt.Sprint("pinned ", id), Date: 10}
		}
		return &tg.MessagesMessages{Messages: []tg.MessageClass{msg(30), msg(7)}}, nil
	})
	if ids := waitPinned(t, s, chat, 2); !slices.Equal(ids, []model.MessageID{7, 30}) {
		t.Fatalf("pinned %v", ids)
	}
	if m, state := s.LookupMessage(chat, 30); state != model.LookupFound || m.Text != "pinned 30" {
		t.Fatalf("the bar's message is %v %+v", state, m)
	}
	ctx := context.Background()
	if err := s.Handle(ctx, &tg.UpdateShort{Update: &tg.UpdatePinnedMessages{Pinned: true, Peer: &tg.PeerUser{UserID: 5}, Messages: []int{12}}}); err != nil {
		t.Fatal(err)
	}
	if err := s.Handle(ctx, &tg.UpdateShort{Update: &tg.UpdatePinnedMessages{Peer: &tg.PeerUser{UserID: 5}, Messages: []int{7}}}); err != nil {
		t.Fatal(err)
	}
	if ids := s.PinnedMessages(chat); !slices.Equal(ids, []model.MessageID{12, 30}) {
		t.Fatalf("after updates pinned %v", ids)
	}
	s.HidePinned(chat)
	if ids := s.PinnedMessages(chat); len(ids) != 0 {
		t.Fatalf("hidden pinned messages show: %v", ids)
	}
	if err := s.Handle(ctx, &tg.UpdateShort{Update: &tg.UpdatePinnedMessages{Pinned: true, Peer: &tg.PeerUser{UserID: 5}, Messages: []int{40}}}); err != nil {
		t.Fatal(err)
	}
	if ids := s.PinnedMessages(chat); !slices.Equal(ids, []model.MessageID{12, 30, 40}) {
		t.Fatalf("a newer pin shows %v", ids)
	}
	if searched != 1 {
		t.Fatalf("searched %d times", searched)
	}
}
