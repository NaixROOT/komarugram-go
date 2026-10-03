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

// A user's photos are Telegram's list, paged past the last photo's ID with
// Telegram's count as the total; a group's are the photo it has and the
// ones its service messages tell it had, the current one first and none
// told twice.
func TestProfilePhotosOfUserAndChannel(t *testing.T) {
	s := testStore(t)
	s.rememberPeers([]tg.UserClass{&tg.User{ID: 5, AccessHash: 6, FirstName: "Ann"}}, []tg.ChatClass{&tg.Channel{ID: 9, Megagroup: true, Title: "group", AccessHash: 7}})
	group := peerID(&tg.PeerChannel{ChannelID: 9})
	var asked []string
	var userPages []*tg.PhotosGetUserPhotosRequest
	var searches []*tg.MessagesSearchRequest
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		asked = append(asked, fmt.Sprintf("%T", in))
		switch r := in.(type) {
		case *tg.PhotosGetUserPhotosRequest:
			userPages = append(userPages, r)
			// Photos 100, 99, … 1: newer photos have greater IDs.
			var photos []tg.PhotoClass
			for id := int64(100); id > 0 && len(photos) < r.Limit; id-- {
				if r.MaxID == 0 || id < r.MaxID {
					photos = append(photos, profilePhoto(id))
				}
			}
			return &tg.PhotosPhotosSlice{Count: 100, Photos: photos}, nil
		case *tg.ChannelsGetFullChannelRequest:
			return &tg.MessagesChatFull{FullChat: &tg.ChannelFull{ChatPhoto: profilePhoto(3)}}, nil
		case *tg.MessagesSearchRequest:
			if _, ok := r.Filter.(*tg.InputMessagesFilterChatPhotos); !ok {
				return nil, fmt.Errorf("filter %T", r.Filter)
			}
			searches = append(searches, r)
			edit := func(id int, p *tg.Photo) tg.MessageClass {
				return &tg.MessageService{ID: id, PeerID: &tg.PeerChannel{ChannelID: 9}, Action: &tg.MessageActionChatEditPhoto{Photo: p}}
			}
			if r.OffsetID == 0 {
				return &tg.MessagesChannelMessages{Count: 3, Messages: []tg.MessageClass{edit(30, profilePhoto(3)), edit(20, profilePhoto(4))}}, nil
			}
			return &tg.MessagesChannelMessages{Count: 3, Messages: []tg.MessageClass{edit(10, profilePhoto(5))}}, nil
		}
		return nil, fmt.Errorf("unexpected %T", in)
	})
	ctx := context.Background()
	user, err := s.ProfilePhotos(ctx, 5, "", 60)
	if err != nil || len(user.Messages) != 60 || user.Total != 100 || !user.More || user.Next == "" {
		t.Fatalf("user's first page: %d photos of %d, more %v %q: %v", len(user.Messages), user.Total, user.More, user.Next, err)
	}
	if user.Messages[0].Key.MessageID != model.ProfilePhotoID(0) || user.Messages[1].Key.MessageID != model.ProfilePhotoID(1) {
		t.Fatalf("ids %v %v", user.Messages[0].Key.MessageID, user.Messages[1].Key.MessageID)
	}
	if user.Messages[0].Media.Width != 640 || len(user.Messages[0].Media.Variants) != 1 || user.Messages[0].Date.Unix() != 1100 {
		t.Fatalf("photo %+v", user.Messages[0])
	}
	loc, ok := s.history.refs[user.Messages[1].Media.ID].input().(*tg.InputPhotoFileLocation)
	if !ok || loc.ID != 99 || loc.ThumbSize != "x" || string(loc.FileReference) != "\x63" {
		t.Fatalf("location %+v", s.history.refs[user.Messages[1].Media.ID].input())
	}
	rest, err := s.ProfilePhotos(ctx, 5, user.Next, 60)
	if err != nil || len(rest.Messages) != 40 || rest.Total != 100 || rest.More || rest.Next != "" {
		t.Fatalf("user's last page: %d photos of %d, more %v %q: %v", len(rest.Messages), rest.Total, rest.More, rest.Next, err)
	}
	if userPages[1].MaxID != 41 || rest.Messages[0].Key.MessageID != model.ProfilePhotoID(60) {
		t.Fatalf("second page asked past %d, starts at %d", userPages[1].MaxID, rest.Messages[0].Key.MessageID)
	}

	channel, err := s.ProfilePhotos(ctx, group, "", 2)
	if err != nil || len(channel.Messages) != 2 || !channel.More || channel.Total != 3 {
		t.Fatalf("group's first page: %d photos of %d, more %v: %v", len(channel.Messages), channel.Total, channel.More, err)
	}
	if channel.Messages[0].Media.ID == channel.Messages[1].Media.ID {
		t.Fatal("a photo told twice")
	}
	older, err := s.ProfilePhotos(ctx, group, channel.Next, 2)
	if err != nil || len(older.Messages) != 1 || older.More || older.Messages[0].Key.MessageID != model.ProfilePhotoID(2) {
		t.Fatalf("group's last page %+v: %v", older, err)
	}
	if searches[1].OffsetID != 20 {
		t.Fatalf("group's second page asked past message %d", searches[1].OffsetID)
	}
	if s.history.refs[older.Messages[0].Media.ID].ProfileMessage != 10 {
		t.Fatal("a group's photo does not know its service message")
	}
	if len(asked) != 5 {
		t.Fatalf("asked %v", asked)
	}
}

// A group's photo whose service message is gone is counted beside
// Telegram's count, on every page.
func TestProfilePhotosCountCurrentWithoutMessage(t *testing.T) {
	s := testStore(t)
	s.rememberPeers(nil, []tg.ChatClass{&tg.Channel{ID: 9, Megagroup: true, Title: "group", AccessHash: 7}})
	group := peerID(&tg.PeerChannel{ChannelID: 9})
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		switch r := in.(type) {
		case *tg.ChannelsGetFullChannelRequest:
			return &tg.MessagesChatFull{FullChat: &tg.ChannelFull{ChatPhoto: profilePhoto(7)}}, nil
		case *tg.MessagesSearchRequest:
			// Messages 40 and 30 tell photos 40 and 30.
			id := 40
			if r.OffsetID != 0 {
				id = r.OffsetID - 10
			}
			edit := &tg.MessageService{ID: id, PeerID: &tg.PeerChannel{ChannelID: 9}, Action: &tg.MessageActionChatEditPhoto{Photo: profilePhoto(int64(id))}}
			return &tg.MessagesChannelMessages{Count: 2, Messages: []tg.MessageClass{edit}}, nil
		}
		return nil, fmt.Errorf("unexpected %T", in)
	})
	first, err := s.ProfilePhotos(context.Background(), group, "", 1)
	if err != nil || len(first.Messages) != 2 || first.Total != 3 {
		t.Fatalf("first page: %d photos of %d: %v", len(first.Messages), first.Total, err)
	}
	second, err := s.ProfilePhotos(context.Background(), group, first.Next, 1)
	if err != nil || second.Total != 3 || len(second.Messages) != 1 || second.Messages[0].Key.MessageID != model.ProfilePhotoID(2) {
		t.Fatalf("second page %+v: %v", second, err)
	}
}

// An expired file reference of a profile photo is renewed for that photo
// alone, as Telegram Desktop does: a user's by its ID, a group's by its
// service message.
func TestProfilePhotoReferenceRenewed(t *testing.T) {
	s := testStore(t)
	s.rememberPeers([]tg.UserClass{&tg.User{ID: 5, AccessHash: 6, FirstName: "Ann"}}, []tg.ChatClass{&tg.Channel{ID: 9, Megagroup: true, Title: "group", AccessHash: 7}})
	group := peerID(&tg.PeerChannel{ChannelID: 9})
	renewed := func(id int64) *tg.Photo {
		p := profilePhoto(id)
		p.FileReference = []byte("new")
		return p
	}
	var asked []string
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		switch r := in.(type) {
		case *tg.PhotosGetUserPhotosRequest:
			asked = append(asked, fmt.Sprintf("user %d %d %d", r.Offset, r.MaxID, r.Limit))
			return &tg.PhotosPhotosSlice{Count: 500, Photos: []tg.PhotoClass{renewed(r.MaxID)}}, nil
		case *tg.ChannelsGetMessagesRequest:
			id := r.ID[0].(*tg.InputMessageID).ID
			asked = append(asked, fmt.Sprintf("message %d", id))
			return &tg.MessagesChannelMessages{Messages: []tg.MessageClass{&tg.MessageService{ID: id, PeerID: &tg.PeerChannel{ChannelID: 9}, Action: &tg.MessageActionChatEditPhoto{Photo: renewed(4)}}}}, nil
		}
		return nil, fmt.Errorf("unexpected %T", in)
	})
	c := s.history
	c.mu.Lock()
	user := c.rememberProfilePhoto(profilePhoto(321), 0)
	chat := c.rememberProfilePhoto(profilePhoto(4), 20)
	c.mu.Unlock()
	ctx := context.Background()
	if err := s.refreshReference(ctx, model.Message{Key: model.MessageKey{ChatID: 5, MessageID: model.ProfilePhotoID(250)}, Media: user}); err != nil {
		t.Fatal(err)
	}
	if err := s.refreshReference(ctx, model.Message{Key: model.MessageKey{ChatID: group, MessageID: model.ProfilePhotoID(7)}, Media: chat}); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(asked) != "[user -1 321 1 message 20]" {
		t.Fatalf("asked %v", asked)
	}
	for _, m := range []*model.MessageMedia{user, &user.Variants[0], chat} {
		if string(s.history.refs[m.ID].Reference) != "new" {
			t.Fatalf("%s keeps the old reference", m.ID)
		}
	}
}
