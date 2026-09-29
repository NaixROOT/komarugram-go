// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/gotd/td/tg"

	"komarugram/internal/crash"
	"komarugram/internal/messenger/model"
)

// The comments to a channel post are a thread of its discussion group: the
// post's copy in the group, and the messages that reply to it there
// (messages.getDiscussionMessage, messages.getReplies). The UI shows a
// thread as a chat of its own, under an id this store makes up for the
// session; everything asked of that chat is asked of the group.
//
// A thread's messages are not saved to the group's history in the cache: they
// are single messages from anywhere in it, and the cache keeps each chat's
// history without gaps.

// threadBase is below every peer id: thread ids count down from it.
const threadBase = -(int64(1) << 62)

// threadPage is how many comments a page asks for.
const threadPage = 50

// thread is the discussion of a channel post.
type thread struct {
	// post is the channel post.
	post model.MessageKey
	// group is the discussion group and top the post's copy there, the
	// thread's root; both 0 until the discussion is found.
	group int64
	top   int
	// root is the root message, album parts included.
	root []model.Message
	// topic is set when the thread is a topic of a forum: its group and
	// root are known from the start, and its messages are those of the
	// topic.
	topic bool
}

func isThread(chat int64) bool { return chat <= threadBase }

// realChat returns the chat a thread chat lives in, 0 while it is not
// known, or chat itself for any other chat.
func (s *Store) realChat(chat int64) int64 {
	if !isThread(chat) {
		return chat
	}
	c := s.history
	c.mu.Lock()
	defer c.mu.Unlock()
	group, _ := c.threadChat(chat)
	return group
}

// threadChat returns the chat the thread chat lives in, and its root; chat
// itself and 0 for any other chat. The caller holds c.mu.
func (c *conversation) threadChat(chat int64) (int64, int) {
	if !isThread(chat) {
		return chat, 0
	}
	if t := c.threads[chat]; t != nil && t.group != 0 {
		return t.group, t.top
	}
	return 0, 0
}

// OpenComments implements model.CommentsStore.
func (s *Store) OpenComments(post model.Message) model.Chat {
	c := s.history
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.threads == nil {
		c.threads, c.threadIDs = map[int64]*thread{}, map[model.MessageKey]int64{}
	}
	chat := model.Chat{Kind: model.KindGroup, Title: c.peers[post.Key.ChatID].Name}
	if id, ok := c.threadIDs[post.Key]; ok {
		chat.ID = id
		if h := c.histories[id]; h != nil && h.Err == nil {
			return chat
		}
	} else {
		chat.ID = threadBase - int64(len(c.threadIDs))
		c.threadIDs[post.Key] = chat.ID
		c.threads[chat.ID] = &thread{post: post.Key}
	}
	if c.closing {
		return chat
	}
	c.histories[chat.ID] = &model.History{LoadingOlder: true, HasOlder: true}
	c.wg.Go(func() {
		defer crash.Recover("comments", func(p *crash.Panic) { s.threadFailed(chat.ID, p) })
		ctx, cancel := context.WithTimeout(c.ctx, time.Minute)
		defer cancel()
		if err := s.openThread(ctx, chat.ID); err != nil {
			s.threadFailed(chat.ID, err)
		}
	})
	return chat
}

func (s *Store) threadFailed(id int64, err error) {
	c := s.history
	c.mu.Lock()
	if h := c.histories[id]; h != nil {
		h.LoadingOlder, h.Err = false, err
		h.Revision++
	}
	c.mu.Unlock()
	s.changed()
}

// openThread finds the discussion of the thread's post and loads its newest
// comments.
func (s *Store) openThread(ctx context.Context, id int64) error {
	c := s.history
	c.mu.Lock()
	api, t := c.api, c.threads[id]
	channel := c.peers[t.post.ChatID]
	c.mu.Unlock()
	if api == nil {
		return errNotConnected
	}
	found, err := api.MessagesGetDiscussionMessage(ctx, &tg.MessagesGetDiscussionMessageRequest{Peer: channel.input(), MsgID: int(t.post.MessageID)})
	if err != nil {
		return err
	}
	s.rememberPeers(found.Users, found.Chats)
	if len(found.Messages) == 0 {
		return errors.New("the post has no discussion")
	}
	peer, ok := messagePeer(found.Messages[0])
	if !ok {
		return errors.New("the discussion has no group")
	}
	group, top := peerID(peer), found.Messages[0].GetID()
	for _, m := range found.Messages {
		top = min(top, m.GetID())
	}
	root, err := s.threadMessages(ctx, found.Messages)
	if err != nil {
		return err
	}
	c.mu.Lock()
	t.group, t.top, t.root = group, top, root
	c.mu.Unlock()
	return s.threadPage(ctx, id, 0)
}

// threadMessages converts messages of a thread without saving them to the
// cache's history.
func (s *Store) threadMessages(ctx context.Context, raw []tg.MessageClass) ([]model.Message, error) {
	c := s.history
	c.apply.Lock()
	defer c.apply.Unlock()
	msgs, err := s.convert(ctx, raw, false, 0)
	sort.Slice(msgs, func(i, j int) bool { return msgs[i].Key.MessageID < msgs[j].Key.MessageID })
	return msgs, err
}

// threadPage loads the comments before offset, the newest ones for 0, and
// puts them in the thread's history, with the root once the first comment
// is in.
func (s *Store) threadPage(ctx context.Context, id int64, offset int) error {
	c := s.history
	c.mu.Lock()
	api, t := c.api, c.threads[id]
	group := c.peers[t.group]
	top := t.top
	c.mu.Unlock()
	if api == nil {
		return errNotConnected
	}
	res, err := api.MessagesGetReplies(ctx, &tg.MessagesGetRepliesRequest{Peer: group.input(), MsgID: top, OffsetID: offset, Limit: threadPage})
	if err != nil {
		return err
	}
	page, ok := res.AsModified()
	if !ok {
		return errors.New("unexpected replies")
	}
	s.rememberPeers(page.GetUsers(), page.GetChats())
	msgs, err := s.threadMessages(ctx, page.GetMessages())
	if err != nil {
		return err
	}
	c.mu.Lock()
	h := c.histories[id]
	if h == nil {
		c.mu.Unlock()
		return nil
	}
	older := len(page.GetMessages()) >= threadPage
	merged := append(msgs, h.Messages...)
	if !older {
		merged = append(append([]model.Message(nil), t.root...), merged...)
	}
	h.Messages = dedupe(merged)
	h.HasOlder, h.LoadingOlder, h.Err = older, false, nil
	h.ThreadRoot = model.MessageID(top)
	h.Revision++
	c.mu.Unlock()
	s.changed()
	return nil
}

// dedupe drops repeated messages from msgs, sorted by id, keeping the
// later copy.
func dedupe(msgs []model.Message) []model.Message {
	sort.SliceStable(msgs, func(i, j int) bool { return msgs[i].Key.MessageID < msgs[j].Key.MessageID })
	out := msgs[:0]
	for _, m := range msgs {
		if n := len(out); n > 0 && out[n-1].Key == m.Key {
			out[n-1] = m
			continue
		}
		out = append(out, m)
	}
	return out
}

// loadOlderComments loads the page of comments before the ones shown.
func (s *Store) loadOlderComments(id int64) {
	c := s.history
	c.mu.Lock()
	h, t := c.histories[id], c.threads[id]
	if c.closing || h == nil || t == nil || t.group == 0 || !h.HasOlder || h.LoadingOlder || len(h.Messages) == 0 {
		c.mu.Unlock()
		return
	}
	h.LoadingOlder = true
	offset := int(h.Messages[0].Key.MessageID)
	c.mu.Unlock()
	c.wg.Go(func() {
		defer crash.Recover("comments", func(p *crash.Panic) { s.threadFailed(id, p) })
		ctx, cancel := context.WithTimeout(c.ctx, time.Minute)
		defer cancel()
		if err := s.threadPage(ctx, id, offset); err != nil {
			s.threadFailed(id, err)
		}
	})
}

// mergeThreads puts a new or edited message of a discussion group into the
// threads it belongs to. The caller holds c.mu.
func (c *conversation) mergeThreads(m model.Message) {
	for id, t := range c.threads {
		if t.group != m.Key.ChatID || t.group == 0 {
			continue
		}
		h := c.histories[id]
		if h == nil {
			continue
		}
		in := int(m.ReplyToTopID) == t.top || int(m.ReplyToMessageID) == t.top
		if t.topic {
			// The General topic's messages have no header to say so.
			in = m.TopicID() == t.top
		}
		found := false
		for i := range h.Messages {
			if h.Messages[i].Key == m.Key {
				h.Messages[i] = m
				found = true
			}
		}
		if !found && in {
			h.Messages = dedupe(append(h.Messages, m))
		}
		if found || in {
			h.Revision++
		}
	}
}

// threadsOf are the thread chats of a discussion group. The caller holds
// c.mu.
func (c *conversation) threadsOf(group int64) []int64 {
	var ids []int64
	for id, t := range c.threads {
		if t.group == group && group != 0 {
			ids = append(ids, id)
		}
	}
	return ids
}
