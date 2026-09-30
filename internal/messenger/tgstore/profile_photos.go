// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"strings"
	"time"

	"komarugram/internal/messenger/model"

	"github.com/gotd/td/tg"
)

// profilePhotoLimit is how many photos of a profile are asked for.
const profilePhotoLimit = 100

// bigAvatarSide is the side of the large picture of a profile photo.
const bigAvatarSide = 640

// ProfilePhoto implements model.ProfilePhotoSource. It reads only local peer
// metadata: the large picture of the photo the chat shows, with the small
// one the chat list has already loaded as its variant, so that the viewer
// has something to show while the large one comes.
func (s *Store) ProfilePhoto(chat int64) (model.Message, bool) {
	small, ok := s.Avatar(chat)
	if !ok || small.Media == nil {
		return model.Message{}, false
	}
	c := s.history
	c.mu.Lock()
	defer c.mu.Unlock()
	ref, ok := c.refs[small.Media.ID]
	if !ok || ref.Peer == nil {
		return model.Message{}, false
	}
	peer := c.peers[chat]
	big := *small.Media
	big.ID = small.Media.ID + "/big"
	big.Thumbnail = nil
	big.Width, big.Height = bigAvatarSide, bigAvatarSide
	big.Variants = []model.MessageMedia{{ID: small.Media.ID, MIMEType: small.Media.MIMEType, Width: small.Media.Width, Height: small.Media.Height}}
	ref.Big = true
	c.refs[big.ID] = ref
	return model.Message{
		Kind:       model.MessagePhoto,
		Key:        model.MessageKey{AccountID: c.account, ChatID: chat, MessageID: model.ProfilePhotoID(0)},
		Media:      &big,
		SenderName: peer.Name,
		NoForwards: peer.Rights.NoForwards,
	}, true
}

// ProfilePhotos implements model.ProfilePhotoSource: a user's photos are
// Telegram's list of them, and a group's or channel's are the photo it has
// now and the ones it had, told by the service messages of their changes.
func (s *Store) ProfilePhotos(ctx context.Context, chat int64) ([]model.Message, error) {
	ctx, api, peer, _, done, err := s.peerOperation(ctx, chat)
	if err != nil {
		return nil, err
	}
	defer done()
	var photos []*tg.Photo
	seen := map[int64]bool{}
	add := func(p tg.PhotoClass) {
		if p, ok := p.(*tg.Photo); ok && !seen[p.ID] {
			seen[p.ID] = true
			photos = append(photos, p)
		}
	}
	switch peer.Kind {
	case "user":
		var user tg.InputUserClass = &tg.InputUser{UserID: peer.ID, AccessHash: peer.Hash}
		if peer.Rights.Self {
			user = &tg.InputUserSelf{}
		}
		res, e := api.PhotosGetUserPhotos(ctx, &tg.PhotosGetUserPhotosRequest{UserID: user, Limit: profilePhotoLimit})
		if e != nil {
			return nil, e
		}
		for _, p := range res.GetPhotos() {
			add(p)
		}
	case "chat", "channel":
		if peer.Kind == "chat" {
			full, e := api.MessagesGetFullChat(ctx, peer.ID)
			if e != nil {
				return nil, e
			}
			if f, ok := full.FullChat.(*tg.ChatFull); ok {
				add(f.ChatPhoto)
			}
		} else {
			full, e := api.ChannelsGetFullChannel(ctx, &tg.InputChannel{ChannelID: peer.ID, AccessHash: peer.Hash})
			if e != nil {
				return nil, e
			}
			if f, ok := full.FullChat.(*tg.ChannelFull); ok {
				add(f.ChatPhoto)
			}
		}
		res, e := api.MessagesSearch(ctx, &tg.MessagesSearchRequest{Peer: peer.input(), Filter: &tg.InputMessagesFilterChatPhotos{}, Limit: profilePhotoLimit})
		if e != nil {
			return nil, e
		}
		if modified, ok := res.AsModified(); ok {
			for _, m := range modified.GetMessages() {
				if service, ok := m.(*tg.MessageService); ok {
					if edit, ok := service.Action.(*tg.MessageActionChatEditPhoto); ok {
						add(edit.Photo)
					}
				}
			}
		}
	default:
		return nil, errors.New("no photos")
	}
	c := s.history
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]model.Message, 0, len(photos))
	for _, p := range photos {
		meta, largest := photoMedia(p)
		if largest == "" {
			continue
		}
		ref := fileLocation{ID: p.ID, Hash: p.AccessHash, Reference: p.FileReference, DC: p.DCID, Photo: true, Thumb: largest}
		c.refs[meta.ID] = ref
		for _, v := range meta.Variants {
			ref.Thumb = v.ID[strings.LastIndexByte(v.ID, '/')+1:]
			c.refs[v.ID] = ref
		}
		out = append(out, model.Message{
			Kind:       model.MessagePhoto,
			Key:        model.MessageKey{AccountID: c.account, ChatID: chat, MessageID: model.ProfilePhotoID(len(out))},
			Media:      meta,
			Date:       time.Unix(int64(p.Date), 0),
			SenderName: peer.Name,
			NoForwards: peer.Rights.NoForwards,
		})
	}
	return out, nil
}
