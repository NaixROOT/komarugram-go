// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

// channelRights are the account's rights in a channel or supergroup.
func channelRights(ch *tg.Channel) peerRights {
	r := peerRights{NoForwards: ch.Noforwards, Broadcast: ch.Broadcast, Delete: ch.Creator, Post: ch.Creator}
	if admin, ok := ch.GetAdminRights(); ok {
		r.Delete = r.Delete || admin.DeleteMessages
		r.Post = r.Post || admin.PostMessages
	}
	r.Muted = ch.Left
	r.Left = ch.Left
	r.JoinToSend = ch.JoinToSend
	if ch.Broadcast {
		r.Muted = r.Muted || !r.Post
	} else if !ch.Creator {
		if banned, ok := ch.GetBannedRights(); ok && banned.SendMessages {
			r.Muted, r.Banned = true, true
		}
		if banned, ok := ch.GetDefaultBannedRights(); ok && banned.SendMessages {
			if _, admin := ch.GetAdminRights(); !admin {
				r.Muted, r.Banned = true, true
			}
		}
	}
	return r
}

// MessageRights implements model.RightsSource, with Telegram Desktop's rules
// (HistoryItem::allowsForward, canDelete and canDeleteForEveryone).
func (s *Store) MessageRights(chat int64, msgs []model.Message) model.MessageRights {
	chat = s.realChat(chat)
	s.history.mu.Lock()
	peer, known := s.history.peers[chat]
	s.history.mu.Unlock()
	if len(msgs) == 0 {
		return model.MessageRights{}
	}
	r := model.MessageRights{Forward: !peer.Rights.NoForwards, Save: !peer.Rights.NoForwards, Delete: known, Revoke: known}
	channel := peer.Kind == "channel"
	r.Everyone = channel
	for _, m := range msgs {
		service := m.Kind == model.MessageService
		if service || m.NoForwards {
			r.Forward = false
		}
		if m.NoForwards {
			r.Save = false
		}
		switch {
		case channel:
			// The first message of a channel is its creation.
			ok := m.Key.MessageID != 1 && (peer.Rights.Delete || (m.Outgoing && !service && (!m.Post || peer.Rights.Post)))
			r.Delete = r.Delete && ok
			r.Revoke = r.Delete
		case peer.Kind == "user":
			// Nothing sent to oneself or to a bot can be revoked.
			r.Revoke = r.Revoke && !peer.Rights.Self && !peer.Rights.Bot
		case peer.Kind == "chat":
			r.Revoke = r.Revoke && (m.Outgoing || peer.Rights.Delete)
		}
	}
	return r
}

// CanSend implements model.RightsSource.
func (s *Store) CanSend(chat int64) bool {
	thread := isThread(chat)
	chat = s.realChat(chat)
	s.history.mu.Lock()
	peer, known := s.history.peers[chat]
	s.history.mu.Unlock()
	muted := peer.Rights.Muted
	if thread {
		// Comments may come from anyone the discussion group does not
		// ban, unless it takes them only from members.
		muted = peer.Rights.Banned || peer.Rights.Left && peer.Rights.JoinToSend
	}
	return known && !muted && s.Freeze() == (model.Freeze{})
}

// channelUsername is the username of a public channel or supergroup: its
// own, or the first active of its collectible ones.
func channelUsername(ch *tg.Channel) string {
	if ch.Username != "" {
		return ch.Username
	}
	for _, u := range ch.Usernames {
		if u.Active {
			return u.Username
		}
	}
	return ""
}

// MessageLink implements model.MessageLinker as Telegram Desktop's
// CopyPostLink does: t.me/username/id for a public channel or supergroup,
// t.me/c/channel/id, for members only, for a private one.
func (s *Store) MessageLink(chat int64, id model.MessageID) (string, bool, bool) {
	chat = s.realChat(chat)
	s.history.mu.Lock()
	peer, known := s.history.peers[chat]
	s.history.mu.Unlock()
	if !known || peer.Kind != "channel" || id <= 0 {
		return "", false, false
	}
	if peer.Username != "" {
		return fmt.Sprintf("https://t.me/%s/%d", peer.Username, id), true, true
	}
	return fmt.Sprintf("https://t.me/c/%d/%d", peer.ID, id), false, true
}

// ForwardMessages implements model.MessageForwarder.
func (s *Store) ForwardMessages(ctx context.Context, from int64, ids []model.MessageID, to int64) error {
	from = s.realChat(from)
	c := s.history
	c.mu.Lock()
	api := c.api
	source, okFrom := c.peers[from]
	target, okTo := c.peers[to]
	c.mu.Unlock()
	if api == nil {
		return errors.New("cannot forward while offline")
	}
	if !okFrom || !okTo {
		return errors.New("unknown chat")
	}
	for len(ids) > 0 {
		batch := ids[:min(len(ids), 100)]
		ids = ids[len(batch):]
		req := &tg.MessagesForwardMessagesRequest{FromPeer: source.input(), ToPeer: target.input()}
		var random [8]byte
		for _, id := range batch {
			if _, err := rand.Read(random[:]); err != nil {
				return err
			}
			req.ID = append(req.ID, int(id))
			req.RandomID = append(req.RandomID, int64(binary.LittleEndian.Uint64(random[:])))
		}
		res, err := api.MessagesForwardMessages(ctx, req)
		if err != nil {
			return fmt.Errorf("forward: %w", err)
		}
		_ = s.Handle(ctx, res)
	}
	return nil
}

var (
	_ model.RightsSource     = (*Store)(nil)
	_ model.MessageForwarder = (*Store)(nil)
)
