// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"path/filepath"
	"testing"

	"komarugram/internal/messenger/model"
)

// TestRecentChatsKept checks that the search history outlives the store,
// in the account's cache.
func TestRecentChatsKept(t *testing.T) {
	path := filepath.Join(t.TempDir(), "h")
	s := New(nil)
	if err := s.Configure(context.Background(), "a", path, nil); err != nil {
		t.Fatal(err)
	}
	s.BumpRecentChat(model.Chat{ID: 1, Title: "one"})
	s.BumpRecentChat(model.Chat{ID: 2, Title: "two", Unread: 4})
	s.BumpRecentChat(model.Chat{ID: 3, Title: "three"})
	s.RemoveRecentChat(1)
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	again := New(nil)
	if err := again.Configure(context.Background(), "a", path, nil); err != nil {
		t.Fatal(err)
	}
	defer again.Close()
	got := again.RecentChats()
	if len(got) != 2 || got[0].ID != 3 || got[1].ID != 2 || got[1].Unread != 0 {
		t.Fatalf("history after reopening: %+v", got)
	}
	again.ClearRecentChats()
	if len(again.RecentChats()) != 0 {
		t.Error("the history was not cleared")
	}
}
