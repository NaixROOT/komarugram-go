// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"

	"komarugram/internal/messenger/model"
)

// SetKeep implements model.KeepStore.
func (s *Store) SetKeep(k model.Keep) {
	if cache := s.Cache(); cache != nil {
		cache.SetKeep(k.Deleted, k.Edits)
	}
}

// MessageEdits implements model.KeepStore.
func (s *Store) MessageEdits(ctx context.Context, msg model.Message) ([]model.Message, error) {
	cache := s.Cache()
	if cache == nil {
		return nil, errors.New("no cache")
	}
	return cache.Edits(ctx, msg.Key.ChatID, int(msg.Key.MessageID))
}

// isBotChat reports whether chat is a private chat with a bot, whose
// deleted messages are not kept, as AyuGram does by default.
func (s *Store) isBotChat(chat int64) bool {
	c := s.history
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.peers[chat].Rights.Bot
}

// deleteFromUpdate applies a deletion Telegram tells of: the messages the
// cache keeps stay in the history marked deleted; the rest go, as
// deleteMessages does. chat is 0 for the ids of every chat but channels.
func (s *Store) deleteFromUpdate(ctx context.Context, chat int64, ids []int) (map[int64]int, error) {
	c := s.history
	cache := s.Cache()
	if cache == nil || !cache.KeepsDeleted() {
		return s.deleteMessages(ctx, chat, ids)
	}
	c.apply.Lock()
	kept, err := cache.MarkDeleted(ctx, chat, ids, s.isBotChat)
	if err != nil {
		c.apply.Unlock()
		return nil, err
	}
	keptIDs := map[int]bool{}
	for _, m := range kept {
		keptIDs[int(m.Key.MessageID)] = true
	}
	c.mu.Lock()
	s.showKept(kept)
	tops := map[int64]int{}
	for _, m := range kept {
		if c.top[m.Key.ChatID] == int(m.Key.MessageID) {
			tops[m.Key.ChatID] = int(m.Key.MessageID)
		}
	}
	c.mu.Unlock()
	c.apply.Unlock()
	var rest []int
	for _, id := range ids {
		if !keptIDs[id] {
			rest = append(rest, id)
		}
	}
	if len(rest) > 0 {
		more, err := s.deleteMessages(ctx, chat, rest)
		if err != nil {
			return nil, err
		}
		for chat, top := range more {
			tops[chat] = top
		}
	}
	return tops, nil
}

// showKept puts messages kept deleted in place of theirs in the loaded
// histories and lookups; no page may bring them back. The caller holds
// c.mu.
func (s *Store) showKept(kept []model.Message) {
	c := s.history
	for _, m := range kept {
		c.deleted[m.Key] = true
		if l := c.lookups[m.Key]; l != nil {
			c.lookups[m.Key] = &lookup{msg: m, state: model.LookupFound}
		}
		for _, shown := range append([]int64{m.Key.ChatID}, c.threadsOf(m.Key.ChatID)...) {
			h := c.histories[shown]
			if h == nil {
				continue
			}
			for i := range h.Messages {
				if h.Messages[i].Key == m.Key {
					h.Messages[i] = m
					h.Revision++
				}
			}
		}
	}
}
