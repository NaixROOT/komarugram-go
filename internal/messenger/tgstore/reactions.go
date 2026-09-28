// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gotd/td/tg"

	"komarugram/internal/crash"
	"komarugram/internal/messenger/historycache"
	"komarugram/internal/messenger/model"
)

// Reactions the account may choose at once, without and with Premium: the
// defaults of the app config keys reactions_user_max_default and
// reactions_user_max_premium, which Telegram Desktop falls back on too.
const (
	reactionsLimit        = 1
	reactionsLimitPremium = 3
)

var errNotConnected = errors.New("not connected to Telegram")

// reactionState is what the store knows of the reactions allowed.
type reactionState struct {
	mu sync.Mutex
	// toggling lets one change of the account's reactions through at a
	// time.
	toggling sync.Mutex
	// all are Telegram's reactions (messages.getAvailableReactions), and
	// premium the emoji among them only Premium may choose.
	all        []model.Reaction
	premium    map[string]bool
	allLoaded  bool
	allLoading bool
	// chats are the reactions each chat allows; loading, the chats asked.
	chats   map[int64]chatReactions
	loading map[int64]bool
}

// chatReactions is a chat's ChatReactions.
type chatReactions struct {
	// all: every reaction, custom emoji too when custom; otherwise some.
	all, custom bool
	some        []model.Reaction
}

// convertReactions reads the reactions of a message.
func convertReactions(reactions tg.MessageReactions) []model.Reaction {
	var out []model.Reaction
	for _, r := range reactions.Results {
		reaction, ok := convertReaction(r.Reaction)
		if !ok {
			continue
		}
		reaction.Count = r.Count
		_, reaction.Chosen = r.GetChosenOrder()
		out = append(out, reaction)
	}
	return out
}

func convertReaction(r tg.ReactionClass) (model.Reaction, bool) {
	switch kind := r.(type) {
	case *tg.ReactionEmoji:
		return model.Reaction{Emoji: kind.Emoticon}, true
	case *tg.ReactionCustomEmoji:
		return model.Reaction{DocumentID: kind.DocumentID}, true
	case *tg.ReactionPaid:
		return model.Reaction{Paid: true}, true
	}
	return model.Reaction{}, false
}

func inputReaction(r model.Reaction) tg.ReactionClass {
	if r.DocumentID != 0 {
		return &tg.ReactionCustomEmoji{DocumentID: r.DocumentID}
	}
	return &tg.ReactionEmoji{Emoticon: r.Emoji}
}

func (s *Store) reactionLimit() int {
	if s.Me().Premium {
		return reactionsLimitPremium
	}
	return reactionsLimit
}

// ChatReactions implements model.Reactor.
func (s *Store) ChatReactions(chat int64) ([]model.Reaction, int, bool) {
	if chat = s.realChat(chat); chat == 0 {
		return nil, 0, false
	}
	r := &s.reactions
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.allLoaded {
		if !r.allLoading {
			r.allLoading = true
			s.goReactions(func(ctx context.Context, api *tg.Client) { s.loadAvailableReactions(ctx, api) })
		}
		return nil, 0, false
	}
	allowed, ok := r.chats[chat]
	if !ok {
		if !r.loading[chat] {
			if r.loading == nil {
				r.loading = map[int64]bool{}
			}
			r.loading[chat] = true
			s.goReactions(func(ctx context.Context, api *tg.Client) { s.loadChatReactions(ctx, api, chat) })
		}
		return nil, 0, false
	}
	premium := s.Me().Premium
	var list []model.Reaction
	if allowed.all {
		for _, one := range r.all {
			if premium || !r.premium[one.Emoji] {
				list = append(list, one)
			}
		}
	} else {
		list = allowed.some
	}
	return list, s.reactionLimit(), true
}

// goReactions runs load with the API, off the frame.
func (s *Store) goReactions(load func(context.Context, *tg.Client)) {
	c := s.history
	c.mu.Lock()
	defer c.mu.Unlock()
	api, ctx := c.api, c.ctx
	if c.closing {
		return
	}
	c.wg.Go(func() {
		defer crash.Recover("reactions", nil)
		if api == nil {
			// Not connected yet: ask again later.
			time.Sleep(time.Second)
			s.forgetReactionLoads()
			s.changed()
			return
		}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		load(ctx, api)
		s.changed()
	})
}

func (s *Store) forgetReactionLoads() {
	r := &s.reactions
	r.mu.Lock()
	r.allLoading = false
	clear(r.loading)
	r.mu.Unlock()
}

func (s *Store) loadAvailableReactions(ctx context.Context, api *tg.Client) {
	res, err := api.MessagesGetAvailableReactions(ctx, 0)
	r := &s.reactions
	r.mu.Lock()
	defer r.mu.Unlock()
	r.allLoading = false
	list, ok := res.(*tg.MessagesAvailableReactions)
	if err != nil || !ok {
		return
	}
	r.all, r.premium = nil, map[string]bool{}
	for _, one := range list.Reactions {
		if one.Inactive {
			continue
		}
		r.all = append(r.all, model.Reaction{Emoji: one.Reaction})
		if one.Premium {
			r.premium[one.Reaction] = true
		}
	}
	r.allLoaded = true
}

func (s *Store) loadChatReactions(ctx context.Context, api *tg.Client, chat int64) {
	c := s.history
	c.mu.Lock()
	peer := c.peers[chat]
	c.mu.Unlock()
	allowed, err := fetchChatReactions(ctx, api, peer)
	r := &s.reactions
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.loading, chat)
	if err != nil {
		return
	}
	if r.chats == nil {
		r.chats = map[int64]chatReactions{}
	}
	r.chats[chat] = allowed
}

// fetchChatReactions asks which reactions peer allows. Private chats allow
// every one.
func fetchChatReactions(ctx context.Context, api *tg.Client, peer peerRecord) (chatReactions, error) {
	var available tg.ChatReactionsClass
	switch peer.Kind {
	case "chat":
		full, err := api.MessagesGetFullChat(ctx, peer.ID)
		if err != nil {
			return chatReactions{}, err
		}
		if f, ok := full.FullChat.(*tg.ChatFull); ok {
			available, _ = f.GetAvailableReactions()
		}
	case "channel":
		full, err := api.ChannelsGetFullChannel(ctx, &tg.InputChannel{ChannelID: peer.ID, AccessHash: peer.Hash})
		if err != nil {
			return chatReactions{}, err
		}
		if f, ok := full.FullChat.(*tg.ChannelFull); ok {
			available, _ = f.GetAvailableReactions()
		}
	default:
		return chatReactions{all: true}, nil
	}
	switch a := available.(type) {
	case *tg.ChatReactionsAll:
		return chatReactions{all: true, custom: a.AllowCustom}, nil
	case *tg.ChatReactionsSome:
		var some []model.Reaction
		for _, one := range a.Reactions {
			if r, ok := convertReaction(one); ok && !r.Paid {
				some = append(some, r)
			}
		}
		return chatReactions{some: some}, nil
	}
	// ChatReactionsNone, or nothing said: no reactions.
	return chatReactions{}, nil
}

// ToggleReaction implements model.Reactor.
func (s *Store) ToggleReaction(msg model.Message, reaction model.Reaction, report func(error)) {
	c := s.history
	chat, id := msg.Key.ChatID, int(msg.Key.MessageID)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing {
		return
	}
	// The change is saved to the cache, which must not hold up a frame.
	c.wg.Go(func() {
		defer crash.Recover("send reaction", func(p *crash.Panic) { report(p) })
		// One change at a time, so that each one starts from the last.
		s.reactions.toggling.Lock()
		defer s.reactions.toggling.Unlock()
		limit := s.reactionLimit()
		var before, after []model.Reaction
		if !s.changeMessage(chat, id, func(m *model.Message) {
			before = m.Reactions
			m.Reactions = model.ToggleReaction(m.Reactions, reaction, limit)
			after = m.Reactions
		}) {
			return
		}
		c.mu.Lock()
		api, ctx, peer := c.api, c.ctx, c.peers[chat]
		c.mu.Unlock()
		err := s.sendReaction(ctx, api, peer, id, model.ChosenReactions(after))
		if err == nil {
			return
		}
		// Undo, unless something changed the reactions since.
		s.changeMessage(chat, id, func(m *model.Message) {
			if sameReactions(m.Reactions, after) {
				m.Reactions = before
			}
		})
		report(err)
	})
}

func (s *Store) sendReaction(ctx context.Context, api *tg.Client, peer peerRecord, id int, chosen []model.Reaction) error {
	if api == nil {
		return errNotConnected
	}
	req := &tg.MessagesSendReactionRequest{Peer: peer.input(), MsgID: id}
	for _, r := range chosen {
		req.Reaction = append(req.Reaction, inputReaction(r))
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	updates, err := api.MessagesSendReaction(ctx, req)
	if err != nil {
		return err
	}
	return s.Handle(ctx, updates)
}

func sameReactions(a, b []model.Reaction) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// applyReactions takes the reactions an update says a message has. A min
// update does not say which are the account's, so those stay as they were.
func (s *Store) applyReactions(chat int64, id int, reactions tg.MessageReactions) {
	next := convertReactions(reactions)
	s.changeMessage(chat, id, func(m *model.Message) {
		if reactions.Min {
			for i := range next {
				next[i].Chosen = false
				for _, old := range m.Reactions {
					if old.Same(next[i]) {
						next[i].Chosen = old.Chosen
					}
				}
			}
		}
		m.Reactions = next
	})
}

// changeMessage changes a message where it is kept: in the open history and
// in the cache. It reports whether the message was found.
func (s *Store) changeMessage(chat int64, id int, change func(*model.Message)) bool {
	c := s.history
	var cache *historycache.Cache
	var ctx context.Context
	// inThread is set when the message is only in a thread, which the
	// cache does not keep.
	inThread := false
	changed := func() *model.Message {
		c.mu.Lock()
		defer c.mu.Unlock()
		cache, ctx = c.cache, c.ctx
		key := model.MessageKey{AccountID: c.account, ChatID: chat, MessageID: model.MessageID(id)}
		var found *model.Message
		for _, shown := range append([]int64{chat}, c.threadsOf(chat)...) {
			h := c.histories[shown]
			if h == nil {
				continue
			}
			for i := range h.Messages {
				if h.Messages[i].Key != key {
					continue
				}
				m := h.Messages[i]
				if found != nil {
					// The same change, once: the thread shows the
					// history's message.
					m = *found
				} else {
					change(&m)
					m.ContentRevision = model.Revision(m)
					found = &m
					inThread = shown != chat
				}
				h.Messages[i] = m
				h.Revision++
			}
		}
		return found
	}()
	if inThread {
		s.changed()
		return true
	}
	if cache == nil {
		s.changed()
		return changed != nil
	}
	c.apply.Lock()
	defer c.apply.Unlock()
	if changed == nil {
		m, ok, err := cache.Message(ctx, chat, id)
		if err != nil || !ok {
			return false
		}
		change(&m)
		m.ContentRevision = model.Revision(m)
		changed = &m
	}
	_ = cache.SaveMessages(ctx, []model.Message{*changed})
	s.changed()
	return true
}
