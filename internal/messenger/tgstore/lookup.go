// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gotd/td/tg"

	"komarugram/internal/crash"
	"komarugram/internal/messenger/model"
)

// lookupRetry is how long a failed lookup waits before it is tried again.
const lookupRetry = 30 * time.Second

// lookup is a single message looked up, as the one a reply quotes.
type lookup struct {
	msg   model.Message
	state model.LookupState
	// retry is when a lookup that failed may be tried again.
	retry time.Time
}

// LookupMessage finds message id of chat that the chat's loaded history does
// not hold: in the cache of its history, then from Telegram
// (channels.getMessages or messages.getMessages). A message from Telegram
// is kept apart from the history, which it would break with a gap, and in
// the cache under its own key for offline use.
func (s *Store) LookupMessage(chat int64, id model.MessageID) (model.Message, model.LookupState) {
	c := s.history
	key := model.MessageKey{AccountID: c.account, ChatID: chat, MessageID: id}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lookups == nil {
		c.lookups = map[model.MessageKey]*lookup{}
	}
	if l := c.lookups[key]; l != nil && (l.retry.IsZero() || time.Now().Before(l.retry)) {
		return l.msg, l.state
	}
	if c.closing || id <= 0 {
		return model.Message{}, model.LookupGone
	}
	c.lookups[key] = &lookup{state: model.LookupLoading}
	c.wg.Go(func() {
		defer crash.Recover("message lookup", func(*crash.Panic) { s.lookedUp(key, model.Message{}, errors.New("panic")) })
		ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
		defer cancel()
		m, err := s.findMessage(ctx, chat, id)
		s.lookedUp(key, m, err)
	})
	return model.Message{}, model.LookupLoading
}

// errMessageGone is a lookup of a message Telegram does not have.
var errMessageGone = errors.New("message gone")

// lookedUp keeps what a lookup found, and tells of it.
func (s *Store) lookedUp(key model.MessageKey, m model.Message, err error) {
	c := s.history
	c.mu.Lock()
	l := &lookup{msg: m, state: model.LookupFound}
	switch {
	case errors.Is(err, errMessageGone):
		l.state = model.LookupGone
	case err != nil:
		// Shown as loading until it is tried again.
		l.state, l.retry = model.LookupLoading, time.Now().Add(lookupRetry)
	}
	c.lookups[key] = l
	c.mu.Unlock()
	s.changed()
}

// findMessage reads message id of chat from the cache, or from Telegram.
func (s *Store) findMessage(ctx context.Context, chat int64, id model.MessageID) (model.Message, error) {
	c := s.history
	c.mu.Lock()
	api, cache, peer, known := c.api, c.cache, c.peers[chat], c.peers[chat].ID != 0
	c.mu.Unlock()
	lookupKey := fmt.Sprintf("lookup/%d/%d", chat, id)
	if cache != nil {
		if m, ok, err := cache.Message(ctx, chat, int(id)); err == nil && ok {
			return m, nil
		}
		var m model.Message
		if ok, err := cache.Get(ctx, lookupKey, &m); err == nil && ok {
			return m, nil
		}
	}
	if api == nil {
		return model.Message{}, errors.New("offline")
	}
	if !known {
		return model.Message{}, errors.New("unknown chat")
	}
	ids := []tg.InputMessageClass{&tg.InputMessageID{ID: int(id)}}
	var res tg.MessagesMessagesClass
	var err error
	if peer.Kind == "channel" {
		res, err = api.ChannelsGetMessages(ctx, &tg.ChannelsGetMessagesRequest{Channel: &tg.InputChannel{ChannelID: peer.ID, AccessHash: peer.Hash}, ID: ids})
	} else {
		res, err = api.MessagesGetMessages(ctx, ids)
	}
	if err != nil {
		return model.Message{}, err
	}
	mod, ok := res.AsModified()
	if !ok {
		return model.Message{}, errMessageGone
	}
	s.rememberPeers(mod.GetUsers(), mod.GetChats())
	c.apply.Lock()
	// Any change an update made is newer than none: start is the most.
	msgs, err := s.convert(ctx, mod.GetMessages(), false, ^uint64(0))
	c.apply.Unlock()
	if err != nil {
		return model.Message{}, err
	}
	for _, m := range msgs {
		if m.Key.MessageID == id && m.Key.ChatID == chat {
			if cache != nil {
				_ = cache.Put(ctx, lookupKey, m)
			}
			return m, nil
		}
	}
	return model.Message{}, errMessageGone
}
