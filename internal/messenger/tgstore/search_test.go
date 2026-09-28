// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

// waitSearch waits for the current search to finish loading.
func waitSearch(t *testing.T, s *Store) model.SearchResults {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for {
		r := s.SearchResults()
		if !r.Loading {
			return r
		}
		if time.Now().After(deadline) {
			t.Fatal("the search did not finish")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestSearchLocal(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	channel := peerID(&tg.PeerChannel{ChannelID: 8})
	s.chats = []model.Chat{
		{ID: 5, Kind: model.KindUser, Title: "Фёдор"},
		{ID: channel, Kind: model.KindChannel, Title: "Новости"},
	}
	_, err := s.ingest(ctx, []tg.MessageClass{
		&tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 5}, Message: "встреча в пятницу", Date: 10},
		&tg.Message{ID: 2, PeerID: &tg.PeerChannel{ChannelID: 8}, Message: "встреча отменена", Date: 20},
	}, false, 0)
	if err != nil {
		t.Fatal(err)
	}

	s.Search(model.SearchQuery{Text: "федор"})
	r := waitSearch(t, s)
	if len(r.Chats) != 1 || r.Chats[0].ID != 5 {
		t.Errorf("chats by name, with Ё as Е: %+v", r.Chats)
	}

	s.Search(model.SearchQuery{Text: "встреч"})
	r = waitSearch(t, s)
	if len(r.Messages) != 2 || r.Messages[0].Message.Key.MessageID != 2 || r.Messages[0].Chat.Title != "Новости" {
		t.Errorf("messages, newest first, with their chats: %+v", r.Messages)
	}

	s.Search(model.SearchQuery{Text: "встреч", Section: model.SearchChannels})
	r = waitSearch(t, s)
	if len(r.Messages) != 1 || r.Messages[0].Chat.ID != channel || len(r.Chats) != 0 {
		t.Errorf("channels only: %+v", r)
	}

	s.Search(model.SearchQuery{Text: "встреч", Global: true})
	if r = waitSearch(t, s); !errors.Is(r.Err, model.ErrSearchOffline) {
		t.Errorf("a global search without Telegram: %v", r.Err)
	}
}

func TestRevealDropsPagesOfTheOldHistory(t *testing.T) {
	s := testStore(t)
	chat := int64(5)
	s.history.histories[chat] = &model.History{}
	s.history.mu.Lock()
	epoch := s.history.epochs[chat]
	s.history.mu.Unlock()

	s.Reveal(chat, 40)
	s.history.histories[chat] = &model.History{}
	old := []model.Message{{Key: model.MessageKey{AccountID: "a", ChatID: chat, MessageID: 900}}}
	s.finishPageAt(epoch, chat, 0, old, true, nil)
	if len(s.History(chat).Messages) != 0 {
		t.Error("a page of the old history was merged into the revealed one")
	}
	if v, ok := s.Viewport(chat); !ok || v.AnchorMessageID != 40 || v.AtEnd {
		t.Errorf("the viewport does not start at the message: %+v", v)
	}
}
