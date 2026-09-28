// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/gotd/td/tg"

	"komarugram/internal/crash"
	"komarugram/internal/messenger/model"
)

// pinnedLimit is how many of the latest pinned messages of a chat are
// loaded: messages.search pages them, and the bar rarely goes further.
const pinnedLimit = 100

// pinned are a chat's pinned messages, as the cache keeps them.
type pinned struct {
	IDs []model.MessageID
	// Hidden is the latest pinned message when the account hid them all;
	// a newer pin shows them again.
	Hidden model.MessageID `json:",omitempty"`
	// loading is set while they are on their way; loaded, once Telegram
	// told them. A load that failed is tried again after retry.
	loading, loaded bool
	retry           time.Time
}

func pinnedKey(chat int64) string { return fmt.Sprintf("pinned/%d", chat) }

// shown are the pinned messages to show: none while they are hidden.
func (p *pinned) shown() []model.MessageID {
	if len(p.IDs) == 0 || p.Hidden != 0 && p.IDs[len(p.IDs)-1] <= p.Hidden {
		return nil
	}
	return p.IDs
}

// PinnedMessages returns chat's pinned messages: from the cache at first,
// then from Telegram (messages.search with the pinned filter), and as
// updates change them.
func (s *Store) PinnedMessages(chat int64) []model.MessageID {
	if isThread(chat) {
		return nil
	}
	c := s.history
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.pinned == nil {
		c.pinned = map[int64]*pinned{}
	}
	p := c.pinned[chat]
	if p == nil {
		p = &pinned{}
		c.pinned[chat] = p
	}
	if !p.loaded && !p.loading && !c.closing && c.cache != nil && time.Now().After(p.retry) {
		p.loading = true
		c.wg.Go(func() {
			defer crash.Recover("pinned messages", func(*crash.Panic) { s.pinnedLoaded(chat, nil, false) })
			ctx, cancel := context.WithTimeout(c.ctx, 30*time.Second)
			defer cancel()
			s.loadPinned(ctx, chat)
		})
	}
	return slices.Clone(p.shown())
}

// HidePinned hides chat's pinned messages until a newer one is pinned.
func (s *Store) HidePinned(chat int64) {
	c := s.history
	c.mu.Lock()
	p := c.pinned[chat]
	if p == nil || len(p.IDs) == 0 {
		c.mu.Unlock()
		return
	}
	p.Hidden = p.IDs[len(p.IDs)-1]
	saved := pinned{IDs: slices.Clone(p.IDs), Hidden: p.Hidden}
	cache := c.cache
	c.mu.Unlock()
	if cache != nil {
		_ = cache.Put(context.Background(), pinnedKey(chat), saved)
	}
	s.changed()
}

// loadPinned reads chat's pinned messages from the cache, then asks
// Telegram for them. The messages found are kept as looked up, for the
// bar to show without asking again.
func (s *Store) loadPinned(ctx context.Context, chat int64) {
	c := s.history
	c.mu.Lock()
	api, cache, peer := c.api, c.cache, c.peers[chat]
	c.mu.Unlock()
	var saved pinned
	if ok, err := cache.Get(ctx, pinnedKey(chat), &saved); err == nil && ok {
		s.pinnedLoaded(chat, &saved, false)
	}
	if api == nil || peer.ID == 0 {
		s.pinnedLoaded(chat, nil, false)
		return
	}
	res, err := api.MessagesSearch(ctx, &tg.MessagesSearchRequest{Peer: peer.input(), Filter: &tg.InputMessagesFilterPinned{}, Limit: pinnedLimit})
	if err != nil {
		s.pinnedLoaded(chat, nil, false)
		return
	}
	mod, ok := res.AsModified()
	if !ok {
		s.pinnedLoaded(chat, nil, false)
		return
	}
	s.rememberPeers(mod.GetUsers(), mod.GetChats())
	c.apply.Lock()
	msgs, err := s.convert(ctx, mod.GetMessages(), false, ^uint64(0))
	c.apply.Unlock()
	if err != nil {
		s.pinnedLoaded(chat, nil, false)
		return
	}
	found := pinned{Hidden: saved.Hidden}
	c.mu.Lock()
	if c.lookups == nil {
		c.lookups = map[model.MessageKey]*lookup{}
	}
	for _, m := range msgs {
		if m.Key.ChatID != chat {
			continue
		}
		found.IDs = append(found.IDs, m.Key.MessageID)
		c.lookups[m.Key] = &lookup{msg: m, state: model.LookupFound}
	}
	c.mu.Unlock()
	slices.Sort(found.IDs)
	found.IDs = slices.Compact(found.IDs)
	_ = cache.Put(ctx, pinnedKey(chat), found)
	s.pinnedLoaded(chat, &found, true)
}

// pinnedLoaded keeps the pinned messages loaded, if any, and ends the
// loading when done.
func (s *Store) pinnedLoaded(chat int64, got *pinned, done bool) {
	c := s.history
	c.mu.Lock()
	p := c.pinned[chat]
	if p == nil {
		p = &pinned{}
		c.pinned[chat] = p
	}
	if got != nil && !p.loaded {
		p.IDs, p.Hidden = got.IDs, got.Hidden
	}
	if done {
		p.loaded = true
	}
	if got == nil || done {
		p.loading = false
	}
	if got == nil && !done {
		p.retry = time.Now().Add(lookupRetry)
	}
	c.mu.Unlock()
	s.changed()
}

// applyPinned applies an update that pins or unpins messages of chat.
func (s *Store) applyPinned(ctx context.Context, chat int64, ids []int, pin bool) {
	c := s.history
	c.mu.Lock()
	p := c.pinned[chat]
	if p == nil {
		// Not asked for yet: they load when the chat opens.
		c.mu.Unlock()
		return
	}
	for _, id := range ids {
		i, found := slices.BinarySearch(p.IDs, model.MessageID(id))
		switch {
		case pin && !found:
			p.IDs = slices.Insert(p.IDs, i, model.MessageID(id))
		case !pin && found:
			p.IDs = slices.Delete(p.IDs, i, i+1)
		}
	}
	saved := pinned{IDs: slices.Clone(p.IDs), Hidden: p.Hidden}
	cache := c.cache
	c.mu.Unlock()
	if cache != nil {
		_ = cache.Put(ctx, pinnedKey(chat), saved)
	}
}
