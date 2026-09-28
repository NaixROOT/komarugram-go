// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// The search in a chat asks messages.search while connected, page after
// page, and the cache's index offline; what it finds is not saved.
func TestSearchChat(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	chat := int64(5)
	s.history.peers[chat] = peerRecord{Kind: "user", ID: 5, Hash: 55}
	if _, err := s.ingest(ctx, []tg.MessageClass{&tg.Message{ID: 3, PeerID: &tg.PeerUser{UserID: 5}, Message: "cached apple pie", Date: 10}}, false, 0); err != nil {
		t.Fatal(err)
	}
	page, err := s.SearchChat(ctx, chat, "apple", "", 2)
	if err != nil || len(page.Messages) != 1 || page.Messages[0].Key.MessageID != 3 || page.Next != "" {
		t.Fatalf("offline found %+v, %v", page, err)
	}
	var offsets []int
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		r, ok := in.(*tg.MessagesSearchRequest)
		if !ok || r.Q != "apple" {
			return nil, fmt.Errorf("unexpected %T %+v", in, in)
		}
		offsets = append(offsets, r.OffsetID)
		msg := func(id int) tg.MessageClass {
			return &tg.Message{ID: id, PeerID: &tg.PeerUser{UserID: 5}, Message: "apple", Date: 10}
		}
		if r.OffsetID == 0 {
			return &tg.MessagesMessagesSlice{Count: 3, Messages: []tg.MessageClass{msg(40), msg(30)}}, nil
		}
		return &tg.MessagesMessagesSlice{Count: 3, Messages: []tg.MessageClass{msg(20)}}, nil
	})
	page, err = s.SearchChat(ctx, chat, " apple ", "", 2)
	if err != nil || len(page.Messages) != 2 || page.Messages[0].Key.MessageID != 40 || page.Count != 3 || page.Next != "r30" {
		t.Fatalf("found %+v, %v", page, err)
	}
	page, err = s.SearchChat(ctx, chat, "apple", page.Next, 2)
	if err != nil || len(page.Messages) != 1 || page.Next != "" || offsets[1] != 30 {
		t.Fatalf("the next page is %+v after %v, %v", page, offsets, err)
	}
	if _, ok, _ := s.history.cache.Message(ctx, chat, 40); ok {
		t.Fatal("a found message joined the history's cache")
	}
}
