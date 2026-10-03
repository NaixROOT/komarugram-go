package mockstore

import (
	"slices"

	"komarugram/internal/messenger/model"
)

// demoReactions are the reactions the demo chats allow.
var demoReactions = []string{"👍", "❤", "🔥", "🥰", "👏", "😁", "🤔", "🤯", "😱", "🤬", "😢", "🎉", "🤩", "🤮", "💩", "🙏", "👌", "🕊", "🤡", "🥱", "🥴", "😍", "🐳", "❤‍🔥", "🌚", "🌭", "💯", "🤣", "⚡", "🍌", "🏆"}

// ChatReactions implements model.Reactor: every demo chat allows the same
// reactions, one at a time.
func (s *Store) ChatReactions(chat int64) ([]model.Reaction, int, bool) {
	out := make([]model.Reaction, len(demoReactions))
	for i, e := range demoReactions {
		out[i] = model.Reaction{Emoji: e}
	}
	return out, 1, true
}

// ReactionRank implements model.ReactionOrderer with the order of the demo
// reactions.
func (s *Store) ReactionRank(r model.Reaction) (int, bool) {
	if r.DocumentID != 0 || r.Paid {
		return 0, false
	}
	i := slices.Index(demoReactions, r.Emoji)
	return i, i >= 0
}

// ToggleReaction implements model.Reactor on the demo history.
func (s *Store) ToggleReaction(msg model.Message, r model.Reaction, report func(error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, ok := s.histories[msg.Key.ChatID]
	if !ok {
		return
	}
	for i := range h.Messages {
		if h.Messages[i].Key == msg.Key {
			m := h.Messages[i]
			m.Reactions = model.ToggleReaction(m.Reactions, r, 1)
			m.ContentRevision++
			h.Messages[i] = m
			h.Revision++
			s.histories[msg.Key.ChatID] = h
			return
		}
	}
}
