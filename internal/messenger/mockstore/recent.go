package mockstore

import "komarugram/internal/messenger/model"

// RecentChats implements model.RecentChats, in memory.
func (s *Store) RecentChats() []model.Chat {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]model.Chat(nil), s.recent...)
}

func (s *Store) BumpRecentChat(c model.Chat) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recent = model.BumpChat(s.recent, c, model.RecentChatsLimit)
}

func (s *Store) RemoveRecentChat(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []model.Chat
	for _, c := range s.recent {
		if c.ID != id {
			out = append(out, c)
		}
	}
	s.recent = out
}

func (s *Store) ClearRecentChats() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recent = nil
}
