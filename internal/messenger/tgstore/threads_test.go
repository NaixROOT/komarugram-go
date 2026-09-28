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

// commentsStore is a store with channel 8, whose post 4 is discussed in
// group 20 under message 100.
func commentsStore(t *testing.T) (*Store, *[]bin.Encoder, *sync.Mutex) {
	t.Helper()
	s := testStore(t)
	s.rememberPeers(nil, []tg.ChatClass{&tg.Channel{ID: 8, AccessHash: 9, Broadcast: true, Title: "news"}})
	var mu sync.Mutex
	var asked []bin.Encoder
	inGroup := func(id int, text string, replyTo int) tg.MessageClass {
		m := &tg.Message{ID: id, PeerID: &tg.PeerChannel{ChannelID: 20}, FromID: &tg.PeerUser{UserID: 5}, Message: text, Date: id}
		if replyTo != 0 {
			m.ReplyTo = &tg.MessageReplyHeader{ReplyToMsgID: replyTo}
		}
		m.SetFlags()
		return m
	}
	group := &tg.Channel{ID: 20, AccessHash: 21, Megagroup: true, Title: "news chat", Left: true}
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
		mu.Lock()
		asked = append(asked, in)
		mu.Unlock()
		switch req := in.(type) {
		case *tg.MessagesGetDiscussionMessageRequest:
			if p := req.Peer.(*tg.InputPeerChannel); p.ChannelID != 8 || req.MsgID != 4 {
				t.Errorf("asked for the discussion of %+v %d", p, req.MsgID)
			}
			out.(*tg.MessagesDiscussionMessage).Messages = []tg.MessageClass{inGroup(100, "post", 0)}
			out.(*tg.MessagesDiscussionMessage).Chats = []tg.ChatClass{group}
		case *tg.MessagesGetRepliesRequest:
			if p := req.Peer.(*tg.InputPeerChannel); p.ChannelID != 20 || p.AccessHash != 21 || req.MsgID != 100 {
				t.Errorf("asked for the replies of %+v %d", p, req.MsgID)
			}
			out.(*tg.MessagesMessagesBox).Messages = &tg.MessagesChannelMessages{Messages: []tg.MessageClass{inGroup(102, "second", 100), inGroup(101, "first", 100)},
				Users: []tg.UserClass{&tg.User{ID: 5, FirstName: "Анна"}}}
		case *tg.MessagesSendMessageRequest:
			out.(*tg.UpdatesBox).Updates = &tg.Updates{}
		default:
			t.Errorf("unexpected request %T", in)
		}
		return nil
	}))
	return s, &asked, &mu
}

func TestOpenComments(t *testing.T) {
	s, asked, mu := commentsStore(t)
	channel := peerID(&tg.PeerChannel{ChannelID: 8})
	post := model.Message{Key: model.MessageKey{AccountID: "a", ChatID: channel, MessageID: 4}, Post: true, CommentsOpen: true}
	chat := s.OpenComments(post)
	if !isThread(chat.ID) || chat.Title != "news" {
		t.Fatalf("comments chat %+v", chat)
	}
	var h model.History
	deadline := time.Now().Add(5 * time.Second)
	for {
		h = s.History(chat.ID)
		if !h.LoadingOlder {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the comments never loaded")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if h.Err != nil || h.HasOlder || len(h.Messages) != 3 || h.ThreadRoot != 100 {
		t.Fatalf("history %+v", h)
	}
	group := peerID(&tg.PeerChannel{ChannelID: 20})
	for i, want := range []string{"post", "first", "second"} {
		if m := h.Messages[i]; m.Text != want || m.Key.ChatID != group {
			t.Errorf("message %d: %q in %d, want %q in the group", i, m.Text, m.Key.ChatID, want)
		}
	}
	if again := s.OpenComments(post); again.ID != chat.ID {
		t.Error("the same post got another comments chat")
	}
	// Not in the cache: the group's history there has no gaps.
	if _, ok, _ := s.Cache().Message(context.Background(), group, 101); ok {
		t.Error("a comment was saved to the group's history")
	}

	// A new comment comes; a message of another thread does not show.
	reply := &tg.Message{ID: 103, PeerID: &tg.PeerChannel{ChannelID: 20}, Message: "third", Date: 103, ReplyTo: &tg.MessageReplyHeader{ReplyToMsgID: 101, ReplyToTopID: 100}}
	reply.ReplyTo.(*tg.MessageReplyHeader).SetFlags()
	other := &tg.Message{ID: 104, PeerID: &tg.PeerChannel{ChannelID: 20}, Message: "elsewhere", Date: 104, ReplyTo: &tg.MessageReplyHeader{ReplyToMsgID: 50}}
	reply.SetFlags()
	other.SetFlags()
	if err := s.Handle(context.Background(), &tg.Updates{Updates: []tg.UpdateClass{&tg.UpdateNewChannelMessage{Message: reply}, &tg.UpdateNewChannelMessage{Message: other}}}); err != nil {
		t.Fatal(err)
	}
	h = s.History(chat.ID)
	if len(h.Messages) != 4 || h.Messages[3].Text != "third" {
		t.Errorf("after the update: %+v", h.Messages)
	}

	// Anyone may comment in a group that does not ask to join.
	if !s.CanSend(chat.ID) {
		t.Error("cannot comment")
	}
	// A comment replies to the root; a reply to a comment names the root
	// as the top.
	for _, replyTo := range []model.MessageID{0, 101} {
		if err := s.Send(context.Background(), chat.ID, model.OutgoingMessage{Text: "hi", RandomID: int64(replyTo) + 1, ReplyTo: replyTo}); err != nil {
			t.Fatal(err)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	var sent []*tg.MessagesSendMessageRequest
	for _, req := range *asked {
		if r, ok := req.(*tg.MessagesSendMessageRequest); ok {
			sent = append(sent, r)
		}
	}
	if len(sent) != 2 {
		t.Fatalf("sent %d messages", len(sent))
	}
	for i, want := range [][2]int{{100, 0}, {101, 100}} {
		r := sent[i].ReplyTo.(*tg.InputReplyToMessage)
		top, _ := r.GetTopMsgID()
		if p := sent[i].Peer.(*tg.InputPeerChannel); p.ChannelID != 20 || r.ReplyToMsgID != want[0] || top != want[1] {
			t.Errorf("comment %d: to %d, reply %d, top %d", i, p.ChannelID, r.ReplyToMsgID, top)
		}
	}
}

// TestCommentReaction reacts to a comment, which the cache does not keep.
func TestCommentReaction(t *testing.T) {
	s, _, _ := commentsStore(t)
	channel := peerID(&tg.PeerChannel{ChannelID: 8})
	chat := s.OpenComments(model.Message{Key: model.MessageKey{AccountID: "a", ChatID: channel, MessageID: 4}})
	for s.History(chat.ID).LoadingOlder {
		time.Sleep(5 * time.Millisecond)
	}
	comment := s.History(chat.ID).Messages[1]
	sent := make(chan *tg.MessagesSendReactionRequest, 1)
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(ctx context.Context, in bin.Encoder, out bin.Decoder) error {
		sent <- in.(*tg.MessagesSendReactionRequest)
		out.(*tg.UpdatesBox).Updates = &tg.Updates{}
		return nil
	}))
	s.ToggleReaction(comment, model.Reaction{Emoji: "🔥"}, func(err error) { t.Error(err) })
	req := <-sent
	if p := req.Peer.(*tg.InputPeerChannel); p.ChannelID != 20 || req.MsgID != 101 {
		t.Errorf("reacted to %+v %d", p, req.MsgID)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		if m := s.History(chat.ID).Messages[1]; len(m.Reactions) == 1 && m.Reactions[0].Chosen {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("comment %+v", s.History(chat.ID).Messages[1])
		}
		time.Sleep(5 * time.Millisecond)
	}
}
