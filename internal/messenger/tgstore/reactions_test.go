// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"komarugram/internal/messenger/model"
)

// reactionStore is a store with channel 8 open, holding one post with ❤ ×2.
func reactionStore(t *testing.T, invoke func(in bin.Encoder, out bin.Decoder) error) (*Store, model.Message) {
	t.Helper()
	s := testStore(t)
	s.rememberPeers(nil, []tg.ChatClass{&tg.Channel{ID: 8, AccessHash: 9, Broadcast: true, Title: "news"}})
	chat := peerID(&tg.PeerChannel{ChannelID: 8})
	raw := &tg.Message{ID: 4, PeerID: &tg.PeerChannel{ChannelID: 8}, Message: "post", Date: 4, Post: true,
		Reactions: tg.MessageReactions{Results: []tg.ReactionCount{{Reaction: &tg.ReactionEmoji{Emoticon: "❤"}, Count: 2}}}}
	// Decoding sets the flags; a message made here has to set them itself.
	raw.SetFlags()
	msgs, err := s.ingest(context.Background(), []tg.MessageClass{raw}, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	s.history.histories[chat] = &model.History{Messages: msgs}
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(ctx context.Context, in bin.Encoder, out bin.Decoder) error { return invoke(in, out) }))
	return s, msgs[0]
}

func waitMessage(t *testing.T, s *Store, chat int64, ok func(model.Message) bool) model.Message {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		h := s.History(chat)
		if len(h.Messages) == 1 && ok(h.Messages[0]) {
			return h.Messages[0]
		}
		if time.Now().After(deadline) {
			t.Fatalf("history %+v", h.Messages)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestToggleReactionSends(t *testing.T) {
	sent := make(chan *tg.MessagesSendReactionRequest, 1)
	s, msg := reactionStore(t, func(in bin.Encoder, out bin.Decoder) error {
		req, ok := in.(*tg.MessagesSendReactionRequest)
		if !ok {
			t.Errorf("unexpected request %T", in)
			return nil
		}
		sent <- req
		// Telegram answers with the message's reactions as it has them.
		out.(*tg.UpdatesBox).Updates = &tg.Updates{Updates: []tg.UpdateClass{&tg.UpdateMessageReactions{
			Peer: &tg.PeerChannel{ChannelID: 8}, MsgID: 4,
			Reactions: tg.MessageReactions{Results: []tg.ReactionCount{
				{Reaction: &tg.ReactionEmoji{Emoticon: "❤"}, Count: 2},
				chosen(tg.ReactionCount{Reaction: &tg.ReactionEmoji{Emoticon: "👍"}, Count: 5}),
			}},
		}}}
		return nil
	})
	var reported error
	s.ToggleReaction(msg, model.Reaction{Emoji: "❤"}, func(err error) { reported = err })
	req := <-sent
	if req.MsgID != 4 || len(req.Reaction) != 1 || req.Reaction[0].(*tg.ReactionEmoji).Emoticon != "❤" {
		t.Errorf("sent %+v", req)
	}
	if p := req.Peer.(*tg.InputPeerChannel); p.ChannelID != 8 || p.AccessHash != 9 {
		t.Errorf("sent to %+v", p)
	}
	// The server's answer wins over the change made at once.
	got := waitMessage(t, s, msg.Key.ChatID, func(m model.Message) bool { return len(m.Reactions) == 2 })
	if got.Reactions[1].Emoji != "👍" || got.Reactions[1].Count != 5 || !got.Reactions[1].Chosen || got.Reactions[0].Chosen {
		t.Errorf("reactions %+v", got.Reactions)
	}
	if reported != nil {
		t.Error(reported)
	}
	cached, ok, err := s.Cache().Message(context.Background(), msg.Key.ChatID, 4)
	if err != nil || !ok || len(cached.Reactions) != 2 {
		t.Errorf("cached %+v, %t, %v", cached.Reactions, ok, err)
	}
}

func TestToggleReactionUndoneOnError(t *testing.T) {
	release := make(chan struct{})
	var once sync.Once
	t.Cleanup(func() { once.Do(func() { close(release) }) })
	s, msg := reactionStore(t, func(in bin.Encoder, out bin.Decoder) error {
		<-release
		return tgerr.New(400, "REACTION_INVALID")
	})
	reported := make(chan error, 1)
	s.ToggleReaction(msg, model.Reaction{Emoji: "❤"}, func(err error) { reported <- err })
	// Chosen at once, before Telegram answers.
	waitMessage(t, s, msg.Key.ChatID, func(m model.Message) bool {
		return len(m.Reactions) == 1 && m.Reactions[0].Chosen && m.Reactions[0].Count == 3
	})
	once.Do(func() { close(release) })
	if err := <-reported; !tgerr.Is(err, "REACTION_INVALID") {
		t.Errorf("reported %v", err)
	}
	waitMessage(t, s, msg.Key.ChatID, func(m model.Message) bool {
		return len(m.Reactions) == 1 && !m.Reactions[0].Chosen && m.Reactions[0].Count == 2
	})
}

// TestMinReactionsKeepChoice checks that an update that does not say which
// reactions are the account's leaves its choice as it was.
func TestMinReactionsKeepChoice(t *testing.T) {
	s, msg := reactionStore(t, func(in bin.Encoder, out bin.Decoder) error { return errors.New("offline") })
	s.changeMessage(msg.Key.ChatID, 4, func(m *model.Message) { m.Reactions[0].Chosen = true })
	err := s.Handle(context.Background(), &tg.Updates{Updates: []tg.UpdateClass{
		&tg.UpdateMessageReactions{Peer: &tg.PeerChannel{ChannelID: 8}, MsgID: 4, Reactions: tg.MessageReactions{Min: true,
			Results: []tg.ReactionCount{{Reaction: &tg.ReactionEmoji{Emoticon: "❤"}, Count: 7}}}},
		&tg.UpdateChannelMessageViews{ChannelID: 8, ID: 4, Views: 120},
	}})
	if err != nil {
		t.Fatal(err)
	}
	got := s.History(msg.Key.ChatID).Messages[0]
	if got.Reactions[0].Count != 7 || !got.Reactions[0].Chosen || got.Views != 120 {
		t.Errorf("got %+v, %d views", got.Reactions, got.Views)
	}
	if got.ContentRevision == msg.ContentRevision {
		t.Error("the revision did not change")
	}
}

func TestChatReactions(t *testing.T) {
	s, msg := reactionStore(t, func(in bin.Encoder, out bin.Decoder) error {
		switch in.(type) {
		case *tg.MessagesGetAvailableReactionsRequest:
			out.(*tg.MessagesAvailableReactionsBox).AvailableReactions = &tg.MessagesAvailableReactions{Reactions: []tg.AvailableReaction{
				{Reaction: "❤"}, {Reaction: "👍"}, {Reaction: "🦄", Premium: true}, {Reaction: "💩", Inactive: true},
			}}
		case *tg.ChannelsGetFullChannelRequest:
			full := &tg.ChannelFull{ID: 8, AvailableReactions: &tg.ChatReactionsAll{}}
			full.SetFlags()
			out.(*tg.MessagesChatFull).FullChat = full
		case *tg.HelpGetConfigRequest:
			// No config: the favorite reaction falls back.
			return errors.New("no config")
		default:
			t.Errorf("unexpected request %T", in)
		}
		return nil
	})
	deadline := time.Now().Add(5 * time.Second)
	for {
		list, limit, ok := s.ChatReactions(msg.Key.ChatID)
		if ok {
			if len(list) != 2 || list[0].Emoji != "❤" || list[1].Emoji != "👍" || limit != 1 {
				t.Errorf("reactions %+v, limit %d", list, limit)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the reactions never loaded")
		}
		time.Sleep(5 * time.Millisecond)
	}
	// Private chats allow every reaction, with nothing to ask.
	s.rememberPeers([]tg.UserClass{&tg.User{ID: 5, FirstName: "Анна"}}, nil)
	for {
		if list, _, ok := s.ChatReactions(5); ok {
			if len(list) != 2 {
				t.Errorf("private chat reactions %+v", list)
			}
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func chosen(r tg.ReactionCount) tg.ReactionCount {
	r.SetChosenOrder(0)
	return r
}

// TestReactionsInChosenOrder checks that the account's reactions are kept in
// the order it chose them, which model.ToggleReaction lets go by.
func TestReactionsInChosenOrder(t *testing.T) {
	order := func(r tg.ReactionCount, n int) tg.ReactionCount { r.SetChosenOrder(n); return r }
	got := convertReactions(tg.MessageReactions{Results: []tg.ReactionCount{
		order(tg.ReactionCount{Reaction: &tg.ReactionEmoji{Emoticon: "❤"}, Count: 9}, 2),
		{Reaction: &tg.ReactionEmoji{Emoticon: "👍"}, Count: 5},
		order(tg.ReactionCount{Reaction: &tg.ReactionEmoji{Emoticon: "🔥"}, Count: 3}, 1),
		{Reaction: &tg.ReactionEmoji{Emoticon: "🎉"}, Count: 1},
	}})
	var emoji []string
	for _, r := range got {
		emoji = append(emoji, r.Emoji)
	}
	if want := []string{"👍", "🎉", "🔥", "❤"}; !slices.Equal(emoji, want) {
		t.Errorf("got %v, want %v", emoji, want)
	}

	// An update that does not say which are the account's keeps their order.
	s, msg := reactionStore(t, func(in bin.Encoder, out bin.Decoder) error { return errors.New("offline") })
	s.changeMessage(msg.Key.ChatID, 4, func(m *model.Message) {
		m.Reactions = []model.Reaction{{Emoji: "🔥", Count: 1, Chosen: true}, {Emoji: "❤", Count: 2, Chosen: true}}
	})
	s.applyReactions(msg.Key.ChatID, 4, tg.MessageReactions{Min: true, Results: []tg.ReactionCount{
		{Reaction: &tg.ReactionEmoji{Emoticon: "❤"}, Count: 4},
		{Reaction: &tg.ReactionEmoji{Emoticon: "👍"}, Count: 3},
		{Reaction: &tg.ReactionEmoji{Emoticon: "🔥"}, Count: 2},
	}})
	emoji = nil
	for _, r := range s.History(msg.Key.ChatID).Messages[0].Reactions {
		emoji = append(emoji, r.Emoji)
	}
	if want := []string{"👍", "🔥", "❤"}; !slices.Equal(emoji, want) {
		t.Errorf("min update: got %v, want %v", emoji, want)
	}
}

// TestReactionRank checks that drawing a history asks Telegram for its list
// of reactions once connected, and not while offline.
func TestReactionRank(t *testing.T) {
	asked := make(chan struct{}, 4)
	s, _ := reactionStore(t, func(in bin.Encoder, out bin.Decoder) error {
		switch in.(type) {
		case *tg.MessagesGetAvailableReactionsRequest:
			asked <- struct{}{}
			out.(*tg.MessagesAvailableReactionsBox).AvailableReactions = &tg.MessagesAvailableReactions{Reactions: []tg.AvailableReaction{
				{Reaction: "👍"}, {Reaction: "🙈", Inactive: true}, {Reaction: "❤"},
			}}
			return nil
		}
		return errors.New("not here")
	})
	s.history.mu.Lock()
	api := s.history.api
	s.history.api = nil
	s.history.mu.Unlock()
	if _, ok := s.ReactionRank(model.Reaction{Emoji: "❤"}); ok {
		t.Fatal("ranked before the list loaded")
	}
	s.reactions.mu.Lock()
	loading := s.reactions.allLoading
	s.reactions.mu.Unlock()
	if loading {
		t.Fatal("asked while offline")
	}
	s.history.mu.Lock()
	s.history.api = api
	s.history.mu.Unlock()
	s.ReactionRank(model.Reaction{Emoji: "❤"})
	deadline := time.Now().Add(5 * time.Second)
	for {
		if _, ok := s.ReactionRank(model.Reaction{Emoji: "❤"}); ok {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the list did not load")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if len(asked) != 1 {
		t.Fatalf("asked %d times", len(asked))
	}
	for emoji, want := range map[string]int{"👍": 0, "🙈": 1, "❤": 2} {
		if got, ok := s.ReactionRank(model.Reaction{Emoji: emoji}); !ok || got != want {
			t.Errorf("%s: rank %d, %t; want %d", emoji, got, ok, want)
		}
	}
	if _, ok := s.ReactionRank(model.Reaction{Emoji: "🍌"}); ok {
		t.Error("ranked a reaction not in the list")
	}
}
