// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"testing"

	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

func TestMessageRights(t *testing.T) {
	s := testStore(t)
	channel := peerID(&tg.PeerChannel{ChannelID: 8})
	group := peerID(&tg.PeerChannel{ChannelID: 9})
	s.rememberPeers([]tg.UserClass{
		&tg.User{ID: 5, FirstName: "Фёдор"},
		&tg.User{ID: 6, Bot: true, FirstName: "bot"},
		&tg.User{ID: 7, Self: true, FirstName: "me"},
	}, []tg.ChatClass{
		&tg.Channel{ID: 8, Broadcast: true, Title: "news"},
		&tg.Channel{ID: 9, Megagroup: true, Title: "group", Noforwards: true},
	})
	msg := func(chat int64, id int, out bool) model.Message {
		return model.Message{Key: model.MessageKey{ChatID: chat, MessageID: model.MessageID(id)}, Outgoing: out}
	}
	for _, c := range []struct {
		name string
		chat int64
		msgs []model.Message
		want model.MessageRights
	}{
		{"private", 5, []model.Message{msg(5, 3, false), msg(5, 4, true)}, model.MessageRights{Forward: true, Save: true, Delete: true, Revoke: true}},
		{"saved", 7, []model.Message{msg(7, 3, true)}, model.MessageRights{Forward: true, Save: true, Delete: true}},
		{"bot", 6, []model.Message{msg(6, 3, true)}, model.MessageRights{Forward: true, Save: true, Delete: true}},
		{"channel, not admin", channel, []model.Message{msg(channel, 3, false)}, model.MessageRights{Forward: true, Save: true, Everyone: true}},
		{"protected group, own message", group, []model.Message{msg(group, 3, true)}, model.MessageRights{Delete: true, Revoke: true, Everyone: true}},
		{"protected group, others' message", group, []model.Message{msg(group, 3, false)}, model.MessageRights{Everyone: true}},
	} {
		if got := s.MessageRights(c.chat, c.msgs); got != c.want {
			t.Errorf("%s: %+v, want %+v", c.name, got, c.want)
		}
	}
	service := msg(5, 8, true)
	service.Kind = model.MessageService
	if r := s.MessageRights(5, []model.Message{service}); r.Forward || !r.Save {
		t.Errorf("a service message: %+v", r)
	}

	s.rememberPeers(nil, []tg.ChatClass{&tg.Channel{ID: 8, Broadcast: true, Title: "news", Creator: true}})
	if r := s.MessageRights(channel, []model.Message{msg(channel, 3, false)}); !r.Delete {
		t.Error("the channel's creator cannot delete")
	}
	if r := s.MessageRights(channel, []model.Message{msg(channel, 1, false)}); r.Delete {
		t.Error("the first message of a channel can be deleted")
	}
	if s.CanSend(group) != true || s.CanSend(channel) != true {
		t.Error("a member of a group or the creator of a channel cannot send")
	}
	s.rememberPeers(nil, []tg.ChatClass{&tg.Channel{ID: 8, Broadcast: true, Title: "news"}})
	if s.CanSend(channel) {
		t.Error("a subscriber can post to a channel")
	}
}

// A user's username is kept, as a channel's is, and a min user without one
// does not clear it.
func TestUserUsername(t *testing.T) {
	s := testStore(t)
	s.rememberPeers([]tg.UserClass{
		&tg.User{ID: 5, FirstName: "Ann", Username: "ann"},
		&tg.User{ID: 6, FirstName: "Bob", Usernames: []tg.Username{{Username: "old"}, {Username: "bob", Active: true}}},
	}, nil)
	s.rememberPeers([]tg.UserClass{&tg.User{ID: 5, Min: true, FirstName: "Ann"}}, nil)
	if got := s.Username(5); got != "ann" {
		t.Errorf("user 5: %q", got)
	}
	if got := s.Username(6); got != "bob" {
		t.Errorf("user 6: %q", got)
	}
}
