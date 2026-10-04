// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"time"

	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

var _ model.NoticeSource = (*Store)(nil)

// noticeAge is how old a message may be and still be told of: older ones
// come with the catch-up after a reconnection, and are not news.
const noticeAge = 10 * time.Minute

func (s *Store) SetNotices(f func(model.MessageNotice)) {
	if f == nil {
		s.notices.Store(nil)
		return
	}
	s.notices.Store(&f)
}

// notice tells of m, a new message of chat that came in raw.
func (s *Store) notice(m model.Message, chat model.Chat, raw tg.MessageClass) {
	f := s.notices.Load()
	if f == nil || chat.ID == 0 || m.Outgoing || m.Kind == model.MessageService || chat.Kind == model.KindSaved || time.Since(m.Date) > noticeAge {
		return
	}
	n := model.MessageNotice{Chat: chat, Message: m.Key.MessageID}
	n.Text, n.Sender = topicPreview(m)
	if chat.Kind != model.KindGroup {
		n.Sender = ""
	}
	if r, ok := raw.(*tg.Message); ok {
		n.Silent = r.Silent
	}
	(*f)(n)
}

// setMuted applies a change of a chat's notification settings.
func (s *Store) setMuted(id int64, settings tg.PeerNotifySettings) {
	until, ok := settings.GetMuteUntil()
	muted := ok && time.Unix(int64(until), 0).After(time.Now())
	s.mu.Lock()
	chats := append([]model.Chat(nil), s.chats...)
	for i := range chats {
		if chats[i].ID == id {
			chats[i].Muted = muted
		}
	}
	s.chats = chats
	s.mu.Unlock()
}
