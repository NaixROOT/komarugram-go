// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// The blocked peers load page after page on the first ask, and follow
// UpdatePeerBlocked.
func TestBlocked(t *testing.T) {
	s := testStore(t)
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		r, ok := in.(*tg.ContactsGetBlockedRequest)
		if !ok {
			return nil, fmt.Errorf("unexpected %T", in)
		}
		var page []tg.PeerBlocked
		if r.Offset == 0 {
			for i := range blockedPage {
				page = append(page, tg.PeerBlocked{PeerID: &tg.PeerUser{UserID: int64(1000 + i)}})
			}
		} else {
			page = []tg.PeerBlocked{{PeerID: &tg.PeerUser{UserID: 7}}}
		}
		return &tg.ContactsBlockedSlice{Count: blockedPage + 1, Blocked: page}, nil
	})
	if s.Blocked(7) {
		t.Fatal("blocked before the list loaded")
	}
	deadline := time.Now().Add(5 * time.Second)
	for !s.Blocked(7) {
		if time.Now().After(deadline) {
			t.Fatal("the second page never came")
		}
		time.Sleep(time.Millisecond)
	}
	if !s.Blocked(1000) || s.Blocked(8) {
		t.Fatal("the list is wrong")
	}
	if err := s.Handle(context.Background(), &tg.UpdateShort{Update: &tg.UpdatePeerBlocked{PeerID: &tg.PeerUser{UserID: 7}}}); err != nil {
		t.Fatal(err)
	}
	if s.Blocked(7) {
		t.Fatal("unblocked still blocked")
	}
}
