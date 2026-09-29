// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"komarugram/internal/messenger/model"
)

// Pinned chats stay at the top of the list, in the order they were pinned
// (messages.toggleDialogPin). Only the main list is read, so the pins of
// other folders and of the archive are not told of here.

// PinChat implements model.ChatListActions.
func (s *Store) PinChat(ctx context.Context, chat int64, pin bool) error {
	c := s.history
	c.mu.Lock()
	api, peer := c.api, c.peers[chat]
	c.mu.Unlock()
	if api == nil || peer.ID == 0 {
		return errNotConnected
	}
	_, err := api.MessagesToggleDialogPin(ctx, &tg.MessagesToggleDialogPinRequest{Pinned: pin, Peer: &tg.InputDialogPeer{Peer: peer.input()}})
	switch {
	case tgerr.Is(err, "PINNED_DIALOGS_TOO_MUCH"):
		return model.ErrPinnedTooMuch
	case err != nil:
		return err
	}
	s.publish(func() { s.chats = model.WithPinned(s.chats, chat, pin) })
	// What the list looks like offline follows.
	_ = s.persistDialogs(ctx)
	return nil
}

// applyPin puts a pin or unpin that an update tells of in the list.
func (s *Store) applyPin(u *tg.UpdateDialogPinned) {
	if folder, ok := u.GetFolderID(); ok && folder != 0 {
		return
	}
	peer, ok := u.Peer.(*tg.DialogPeer)
	if !ok {
		return
	}
	id := peerID(peer.Peer)
	s.mu.Lock()
	s.chats = model.WithPinned(s.chats, id, u.Pinned)
	s.mu.Unlock()
}

// applyPinOrder puts the order of the pinned chats that an update tells of in
// the list; without the order, the update only says that it changed, and the
// pinned chats are read again.
func (s *Store) applyPinOrder(ctx context.Context, u *tg.UpdatePinnedDialogs) {
	if folder, ok := u.GetFolderID(); ok && folder != 0 {
		return
	}
	order, ok := u.GetOrder()
	if !ok {
		s.readPinned(ctx)
		return
	}
	var ids []int64
	for _, p := range order {
		if p, ok := p.(*tg.DialogPeer); ok {
			ids = append(ids, peerID(p.Peer))
		}
	}
	s.mu.Lock()
	s.chats = model.WithPinOrder(s.chats, ids)
	s.mu.Unlock()
}

// readPinned reads which chats are pinned, and in what order.
func (s *Store) readPinned(ctx context.Context) {
	c := s.history
	c.mu.Lock()
	api := c.api
	c.mu.Unlock()
	if api == nil {
		return
	}
	res, err := api.MessagesGetPinnedDialogs(ctx, 0)
	if err != nil {
		return
	}
	var ids []int64
	for _, d := range res.Dialogs {
		if d, ok := d.(*tg.Dialog); ok && d.Pinned {
			ids = append(ids, peerID(d.Peer))
		}
	}
	s.mu.Lock()
	s.chats = model.WithPinOrder(s.chats, ids)
	s.mu.Unlock()
}

// MarkChatRead implements model.ChatListActions: everything up to the chat's
// last message is read, whatever Ghost says.
func (s *Store) MarkChatRead(chat int64) {
	c := s.history
	c.mu.Lock()
	top := c.top[chat]
	c.mu.Unlock()
	if top > 0 {
		s.MarkRead(chat, model.MessageID(top), true)
	}
}
