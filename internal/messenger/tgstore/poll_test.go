// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

// TestPollLeftChannel checks that a channel the account is not in is kept up
// to date while it is on screen, and left alone once it is not.
func TestPollLeftChannel(t *testing.T) {
	pollFirst, pollDefault = 10*time.Millisecond, 10*time.Millisecond
	t.Cleanup(func() { pollFirst, pollDefault = time.Second, time.Second })
	s := testStore(t)
	chat := peerID(&tg.PeerChannel{ChannelID: 8})
	s.rememberPeers(nil, []tg.ChatClass{&tg.Channel{ID: 8, AccessHash: 9, Broadcast: true, Left: true, Title: "news"}})
	s.history.histories[chat] = &model.History{}

	var mu sync.Mutex
	var asked []int
	post := func(id int) tg.MessageClass {
		return &tg.Message{ID: id, PeerID: &tg.PeerChannel{ChannelID: 8}, Message: "post", Date: id, Post: true}
	}
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
		switch req := in.(type) {
		case *tg.ChannelsGetFullChannelRequest:
			out.(*tg.MessagesChatFull).FullChat = &tg.ChannelFull{ID: 8, Pts: 10}
		case *tg.UpdatesGetChannelDifferenceRequest:
			if ch := req.Channel.(*tg.InputChannel); ch.ChannelID != 8 || ch.AccessHash != 9 {
				t.Errorf("asked for channel %+v", ch)
			}
			mu.Lock()
			asked = append(asked, req.Pts)
			mu.Unlock()
			box := out.(*tg.UpdatesChannelDifferenceBox)
			switch req.Pts {
			case 10:
				// Not the whole difference: the rest is asked for at once.
				box.ChannelDifference = &tg.UpdatesChannelDifference{Pts: 11, NewMessages: []tg.MessageClass{post(5)}}
			case 11:
				box.ChannelDifference = &tg.UpdatesChannelDifference{Final: true, Pts: 12, NewMessages: []tg.MessageClass{post(6)},
					OtherUpdates: []tg.UpdateClass{&tg.UpdateEditChannelMessage{Message: &tg.Message{ID: 5, PeerID: &tg.PeerChannel{ChannelID: 8}, Message: "edited", Date: 5, EditDate: 7, Post: true}}}}
			default:
				box.ChannelDifference = &tg.UpdatesChannelDifferenceEmpty{Final: true, Pts: req.Pts}
			}
		default:
			t.Errorf("unexpected request %T", in)
		}
		return nil
	}))

	viewer := new(int)
	s.WatchChat(viewer, chat)
	deadline := time.Now().Add(5 * time.Second)
	for {
		h := s.History(chat)
		mu.Lock()
		polled := len(asked)
		mu.Unlock()
		// After the whole difference, polling goes on at the server's pace.
		if len(h.Messages) == 2 && h.Messages[0].Text == "edited" && polled >= 3 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("history %+v", h.Messages)
		}
		time.Sleep(10 * time.Millisecond)
	}
	for _, c := range s.Chats() {
		if c.ID == chat {
			t.Error("a channel the account is not in was added to its chats")
		}
	}

	s.WatchChat(viewer, 0)
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	stopped := len(asked)
	mu.Unlock()
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(asked) != stopped {
		t.Errorf("polling went on after the channel was closed: %v", asked)
	}
	if len(asked) < 3 || asked[0] != 10 || asked[1] != 11 || asked[2] != 12 {
		t.Errorf("asked for pts %v, want 10, 11, 12…", asked)
	}
}

// TestNoPollForJoinedChannel checks that channels the account is in are left
// to the updates manager.
func TestNoPollForJoinedChannel(t *testing.T) {
	pollFirst = 10 * time.Millisecond
	t.Cleanup(func() { pollFirst = time.Second })
	s := testStore(t)
	chat := peerID(&tg.PeerChannel{ChannelID: 8})
	s.rememberPeers(nil, []tg.ChatClass{&tg.Channel{ID: 8, AccessHash: 9, Broadcast: true, Title: "news"}})
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
		t.Errorf("a joined channel was polled: %T", in)
		return nil
	}))
	s.WatchChat(new(int), chat)
	s.WatchChat(new(int), 5)
	time.Sleep(100 * time.Millisecond)
}
