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

// Who reacted comes with names; the list of a message follows its
// reactions' can_see_list; the quick reaction is the config's default.
func TestReactedAndQuickReaction(t *testing.T) {
	s := testStore(t)
	chat := int64(-5)
	s.history.peers[chat] = peerRecord{Kind: "chat", ID: 5}
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		switch r := in.(type) {
		case *tg.MessagesGetMessageReactionsListRequest:
			if r.ID != 3 || r.Offset != "next" {
				return nil, fmt.Errorf("asked %+v", r)
			}
			if e, ok := r.Reaction.(*tg.ReactionEmoji); !ok || e.Emoticon != "❤" {
				return nil, fmt.Errorf("reaction %+v", r.Reaction)
			}
			return &tg.MessagesMessageReactionsList{Count: 1, Reactions: []tg.MessagePeerReaction{{PeerID: &tg.PeerUser{UserID: 7}, Reaction: &tg.ReactionEmoji{Emoticon: "❤"}, Date: 5}}, Users: []tg.UserClass{&tg.User{ID: 7, FirstName: "Ann"}}}, nil
		case *tg.HelpGetConfigRequest:
			cfg := &tg.Config{}
			cfg.SetReactionsDefault(&tg.ReactionEmoji{Emoticon: "🔥"})
			return cfg, nil
		case *tg.MessagesGetAvailableReactionsRequest:
			reaction := func(e string) tg.AvailableReaction {
				none := &tg.DocumentEmpty{}
				return tg.AvailableReaction{Reaction: e, StaticIcon: none, AppearAnimation: none, SelectAnimation: none, ActivateAnimation: none, EffectAnimation: none}
			}
			return &tg.MessagesAvailableReactions{Reactions: []tg.AvailableReaction{reaction("👍"), reaction("🔥")}}, nil
		case *tg.MessagesGetFullChatRequest:
			full := &tg.ChatFull{ID: 5, Participants: &tg.ChatParticipantsForbidden{ChatID: 5}}
			full.SetAvailableReactions(&tg.ChatReactionsAll{})
			return &tg.MessagesChatFull{FullChat: full}, nil
		}
		return nil, fmt.Errorf("unexpected %T", in)
	})
	page, err := s.Reacted(context.Background(), model.Message{Key: model.MessageKey{ChatID: chat, MessageID: 3}}, model.Reaction{Emoji: "❤"}, "next", 50)
	if err != nil || len(page.List) != 1 || page.List[0].Name != "Ann" || page.List[0].PeerID != 7 {
		t.Fatalf("listed %+v, %v", page, err)
	}
	var quick model.Reaction
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var ok bool
		if quick, ok = s.QuickReaction(chat); ok {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if quick.Emoji != "🔥" {
		t.Fatalf("the quick reaction is %+v", quick)
	}
	reactions := tg.MessageReactions{CanSeeList: true, Results: []tg.ReactionCount{{Reaction: &tg.ReactionEmoji{Emoticon: "❤"}, Count: 1}}}
	m, _ := convertMessage("a", &tg.Message{ID: 3, PeerID: &tg.PeerChat{ChatID: 5}, Reactions: reactions, Flags: 1 << 20}, nil)
	if !m.ReactionsListed {
		t.Fatal("a message whose reactions may be listed does not say so")
	}
}
