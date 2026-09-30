// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"

	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

// ReadContents implements model.ContentReader, as Telegram Desktop marks a
// voice message listened to when it plays (ApiWrap::markContentsRead):
// messages.readMessageContents, or channels.readMessageContents in a
// channel. The message loses its mark here in any case; Telegram is told
// only with Ghost.SendRead, as AyuGram tells it.
func (s *Store) ReadContents(m model.Message) {
	chat, id := m.Key.ChatID, int(m.Key.MessageID)
	if !m.MediaUnread || m.Outgoing || id <= 0 {
		return
	}
	s.changeMessage(chat, id, func(m *model.Message) { m.MediaUnread = false })
	if !s.ghostSettings().SendRead {
		return
	}
	c := s.history
	c.mu.Lock()
	peer := c.peers[chat]
	c.mu.Unlock()
	if peer.ID == 0 {
		return
	}
	s.goSend("read contents", func(ctx context.Context, api *tg.Client) error {
		if peer.Kind == "channel" {
			_, err := api.ChannelsReadMessageContents(ctx, &tg.ChannelsReadMessageContentsRequest{Channel: &tg.InputChannel{ChannelID: peer.ID, AccessHash: peer.Hash}, ID: []int{id}})
			return err
		}
		_, err := api.MessagesReadMessageContents(ctx, []int{id})
		return err
	})
}

// contentsRead takes the mark off the messages ids of chat that were
// listened to, here or by who they were sent to. Outside channels Telegram
// names no chat, 0 here: the ids are the account's own there, and the
// message is in one of the chats that are not channels.
func (s *Store) contentsRead(chat int64, ids []int) {
	chats := []int64{chat}
	if chat == 0 {
		chats = chats[:0]
		c := s.history
		c.mu.Lock()
		for id := range c.histories {
			if p, ok := c.peers[id]; ok && p.Kind != "channel" {
				chats = append(chats, id)
			}
		}
		c.mu.Unlock()
	}
	for _, chat := range chats {
		for _, id := range ids {
			s.markListened(chat, id)
		}
	}
}

// markListened takes the mark off message id of chat where it is shown. A
// message only the cache has keeps the mark until it is loaded again: the
// cache is not searched for ids that name no chat.
func (s *Store) markListened(chat int64, id int) {
	c := s.history
	c.mu.Lock()
	marked := false
	if h := c.histories[chat]; h != nil {
		for _, m := range h.Messages {
			if int(m.Key.MessageID) == id {
				marked = m.MediaUnread
				break
			}
		}
	}
	c.mu.Unlock()
	if marked {
		s.changeMessage(chat, id, func(m *model.Message) { m.MediaUnread = false })
	}
}

// KeepWaveform implements model.WaveformKeeper: the waveform worked out
// for a voice message sent without one stays with the message, in the
// cache too.
func (s *Store) KeepWaveform(m model.Message, waveform []byte) {
	if m.Media == nil || len(waveform) == 0 {
		return
	}
	s.changeMessage(m.Key.ChatID, int(m.Key.MessageID), func(m *model.Message) {
		if m.Media == nil || len(m.Media.Waveform) > 0 {
			return
		}
		// The media is shared with the copies of the message handed out.
		media := *m.Media
		media.Waveform = waveform
		m.Media = &media
	})
}

var (
	_ model.ContentReader  = (*Store)(nil)
	_ model.WaveformKeeper = (*Store)(nil)
)
