package mockstore

import (
	"strings"

	"komarugram/internal/messenger/model"
)

// Search implements model.Searcher over the demo chats' names: there is no
// Telegram to ask, so both tabs find the same, and no messages.
func (s *Store) Search(q model.SearchQuery) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.query = q
}

func (s *Store) SearchMore()       {}
func (s *Store) SpendPostsSearch() {}

func (s *Store) SearchResults() model.SearchResults {
	s.mu.Lock()
	q := s.query
	s.mu.Unlock()
	r := model.SearchResults{Query: q}
	text := strings.ToLower(strings.TrimSpace(q.Text))
	if text == "" || (q.Section != model.SearchChats && q.Section != model.SearchChannels) {
		return r
	}
	for _, c := range s.Chats() {
		if q.Section == model.SearchChannels && c.Kind != model.KindChannel {
			continue
		}
		if strings.Contains(strings.ToLower(c.Title), text) {
			r.Chats = append(r.Chats, c)
		}
	}
	return r
}
