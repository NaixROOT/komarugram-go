// SPDX-License-Identifier: Unlicense OR MIT

package mockstore

import "komarugram/internal/messenger/model"

// ReadContents implements model.ContentReader on the demo history.
func (s *Store) ReadContents(msg model.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	h, ok := s.histories[msg.Key.ChatID]
	if !ok {
		return
	}
	for i := range h.Messages {
		if h.Messages[i].Key == msg.Key && h.Messages[i].MediaUnread {
			m := h.Messages[i]
			m.MediaUnread = false
			m.ContentRevision++
			h.Messages[i] = m
			h.Revision++
			s.histories[msg.Key.ChatID] = h
			return
		}
	}
}

var _ model.ContentReader = (*Store)(nil)
