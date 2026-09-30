// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"testing"

	"komarugram/internal/messenger/model"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

func profilePhoto(id int64) *tg.Photo {
	return &tg.Photo{ID: id, AccessHash: id * 3, FileReference: []byte{byte(id)}, Date: int(1000 + id), DCID: 2, Sizes: []tg.PhotoSizeClass{
		&tg.PhotoSize{Type: "m", W: 160, H: 160, Size: 100},
		&tg.PhotoSize{Type: "x", W: 640, H: 640, Size: 900},
	}}
}

// The photo a chat shows is known without asking Telegram: its large
// picture is what the viewer downloads, with the small one of the chat list
// as its variant.
func TestProfilePhotoIsTheLargePicture(t *testing.T) {
	s := testStore(t)
	s.rememberPeers([]tg.UserClass{&tg.User{ID: 42, AccessHash: 55, FirstName: "Ann", Photo: &tg.UserProfilePhoto{PhotoID: 9, DCID: 2}}, &tg.User{ID: 43}}, nil)
	m, ok := s.ProfilePhoto(42)
	if !ok || m.Kind != model.MessagePhoto || !model.IsProfilePhoto(m.Key.MessageID) || m.Key.ChatID != 42 {
		t.Fatalf("profile photo %+v %v", m, ok)
	}
	big, ok := s.history.refs[m.Media.ID].input().(*tg.InputPeerPhotoFileLocation)
	if !ok || !big.Big || big.PhotoID != 9 {
		t.Fatalf("large picture's location %+v", s.history.refs[m.Media.ID].input())
	}
	if len(m.Media.Variants) != 1 {
		t.Fatalf("variants %+v", m.Media.Variants)
	}
	small, ok := s.history.refs[m.Media.Variants[0].ID].input().(*tg.InputPeerPhotoFileLocation)
	if !ok || small.Big {
		t.Fatalf("small picture's location %+v", s.history.refs[m.Media.Variants[0].ID].input())
	}
	if _, ok := s.ProfilePhoto(43); ok {
		t.Fatal("a photo of a user without one")
	}
}

// A user's photos are Telegram's list, a group's the photo it has and the
// ones its service messages tell it had; the current one comes first and
// none is told twice.
func TestProfilePhotosOfUserAndChannel(t *testing.T) {
	s := testStore(t)
	s.rememberPeers([]tg.UserClass{&tg.User{ID: 5, AccessHash: 6, FirstName: "Ann"}}, []tg.ChatClass{&tg.Channel{ID: 9, Megagroup: true, Title: "group", AccessHash: 7}})
	group := peerID(&tg.PeerChannel{ChannelID: 9})
	var asked []string
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		asked = append(asked, fmt.Sprintf("%T", in))
		switch r := in.(type) {
		case *tg.PhotosGetUserPhotosRequest:
			return &tg.PhotosPhotos{Photos: []tg.PhotoClass{profilePhoto(1), profilePhoto(2)}}, nil
		case *tg.ChannelsGetFullChannelRequest:
			return &tg.MessagesChatFull{FullChat: &tg.ChannelFull{ChatPhoto: profilePhoto(3)}}, nil
		case *tg.MessagesSearchRequest:
			if _, ok := r.Filter.(*tg.InputMessagesFilterChatPhotos); !ok {
				return nil, fmt.Errorf("filter %T", r.Filter)
			}
			edit := func(id int, p *tg.Photo) tg.MessageClass {
				return &tg.MessageService{ID: id, PeerID: &tg.PeerChannel{ChannelID: 9}, Action: &tg.MessageActionChatEditPhoto{Photo: p}}
			}
			return &tg.MessagesMessages{Messages: []tg.MessageClass{edit(20, profilePhoto(3)), edit(10, profilePhoto(4))}}, nil
		}
		return nil, fmt.Errorf("unexpected %T", in)
	})
	ctx := context.Background()
	user, err := s.ProfilePhotos(ctx, 5)
	if err != nil || len(user) != 2 {
		t.Fatalf("user's photos %d: %v", len(user), err)
	}
	if user[0].Key.MessageID != model.ProfilePhotoID(0) || user[1].Key.MessageID != model.ProfilePhotoID(1) {
		t.Fatalf("ids %v %v", user[0].Key.MessageID, user[1].Key.MessageID)
	}
	if user[0].Media.Width != 640 || len(user[0].Media.Variants) != 1 || user[0].Date.Unix() != 1001 {
		t.Fatalf("photo %+v", user[0])
	}
	loc, ok := s.history.refs[user[1].Media.ID].input().(*tg.InputPhotoFileLocation)
	if !ok || loc.ID != 2 || loc.ThumbSize != "x" || string(loc.FileReference) != "\x02" {
		t.Fatalf("location %+v", s.history.refs[user[1].Media.ID].input())
	}
	channel, err := s.ProfilePhotos(ctx, group)
	if err != nil || len(channel) != 2 {
		t.Fatalf("group's photos %d: %v", len(channel), err)
	}
	if channel[0].Media.ID == channel[1].Media.ID {
		t.Fatal("a photo told twice")
	}
	if len(asked) != 3 {
		t.Fatalf("asked %v", asked)
	}
}
