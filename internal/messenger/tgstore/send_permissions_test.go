package tgstore

import (
	"context"
	"fmt"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"komarugram/internal/messenger/model"
	"testing"
	"time"
)

func TestComposerPermissions(t *testing.T) {
	s := testStore(t)
	id := peerID(&tg.PeerChannel{ChannelID: 8})
	ch := &tg.Channel{ID: 8, Megagroup: true}
	ch.SetDefaultBannedRights(tg.ChatBannedRights{SendPhotos: true, SendGifs: true})
	ch.SetBannedRights(tg.ChatBannedRights{SendPlain: true, SendVoices: true, UntilDate: int(time.Now().Add(time.Hour).Unix())})
	s.rememberPeers(nil, []tg.ChatClass{ch})
	p := s.SendPermissions(id)
	if p.Allows(model.SendText) || p.Allows(model.SendPhoto) || p.Allows(model.SendVoice) || p.Allows(model.SendGIF) || !p.Allows(model.SendSticker) || !p.Allows(model.SendFile) || p.Expiry(model.SendText).IsZero() {
		t.Fatalf("separate bans lost: %+v", p)
	}
	ch.SetAdminRights(tg.ChatAdminRights{DeleteMessages: true})
	s.rememberPeers(nil, []tg.ChatClass{ch})
	if !s.SendPermissions(id).Allows(model.SendAll) {
		t.Fatal("admin subject to member restrictions")
	}
	ch.Broadcast = true
	ch.Megagroup = false
	s.rememberPeers(nil, []tg.ChatClass{ch})
	if s.SendPermissions(id).Any(model.SendAll) {
		t.Fatal("admin without post_messages can publish")
	}
	ch.SetAdminRights(tg.ChatAdminRights{PostMessages: true})
	s.rememberPeers(nil, []tg.ChatClass{ch})
	if !s.SendPermissions(id).Allows(model.SendText) {
		t.Fatal("publisher cannot publish")
	}
}
func TestComposerPermissionExpiryAndDefaultUpdate(t *testing.T) {
	s := testStore(t)
	ch := &tg.Channel{ID: 8, Megagroup: true}
	ch.SetBannedRights(tg.ChatBannedRights{SendPlain: true, UntilDate: int(time.Now().Add(-time.Minute).Unix())})
	s.rememberPeers(nil, []tg.ChatClass{ch})
	id := peerID(&tg.PeerChannel{ChannelID: 8})
	if !s.SendPermissions(id).Allows(model.SendText) {
		t.Fatal("expired restriction still active")
	}
	err := s.Handle(context.Background(), &tg.UpdateShort{Update: &tg.UpdateChatDefaultBannedRights{Peer: &tg.PeerChannel{ChannelID: 8}, DefaultBannedRights: tg.ChatBannedRights{SendPlain: true}}})
	if err != nil {
		t.Fatal(err)
	}
	if s.SendPermissions(id).Allows(model.SendText) {
		t.Fatal("default rights update ignored")
	}
	if err := s.Send(context.Background(), id, model.OutgoingMessage{RandomID: 1, Text: "blocked"}); err == nil {
		t.Fatal("store accepted forbidden text")
	}
}
func TestComposerUnknownAndMinRights(t *testing.T) {
	s := testStore(t)
	id := peerID(&tg.PeerChannel{ChannelID: 8})
	s.history.peers[id] = peerRecord{ID: 8, Kind: "channel"}
	if s.SendPermissions(id).Any(model.SendAll) {
		t.Fatal("old cache grants rights")
	}
	ch := &tg.Channel{ID: 8, Megagroup: true}
	ch.SetBannedRights(tg.ChatBannedRights{SendStickers: true})
	s.rememberPeers(nil, []tg.ChatClass{ch})
	min := &tg.Channel{ID: 8, Min: true, Megagroup: true}
	min.SetDefaultBannedRights(tg.ChatBannedRights{SendPhotos: true})
	s.rememberPeers(nil, []tg.ChatClass{min})
	p := s.SendPermissions(id)
	if p.Allows(model.SendSticker) || p.Allows(model.SendPhoto) || !p.Allows(model.SendText) {
		t.Fatalf("min replaced personal permissions: %+v", p)
	}
}

func TestComposerEachMediaBan(t *testing.T) {
	for _, tc := range []struct {
		name string
		ban  tg.ChatBannedRights
		kind model.SendKind
	}{
		{"photo", tg.ChatBannedRights{SendPhotos: true}, model.SendPhoto},
		{"video", tg.ChatBannedRights{SendVideos: true}, model.SendVideo},
		{"music", tg.ChatBannedRights{SendAudios: true}, model.SendMusic},
		{"file", tg.ChatBannedRights{SendDocs: true}, model.SendFile},
		{"voice", tg.ChatBannedRights{SendVoices: true}, model.SendVoice},
		{"round", tg.ChatBannedRights{SendRoundvideos: true}, model.SendRoundVideo},
		{"sticker", tg.ChatBannedRights{SendStickers: true}, model.SendSticker},
		{"gif", tg.ChatBannedRights{SendGifs: true}, model.SendGIF},
		{"inline", tg.ChatBannedRights{SendInline: true}, model.SendInline},
		{"text", tg.ChatBannedRights{SendPlain: true}, model.SendText},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := testStore(t)
			ch := &tg.Channel{ID: 8, Megagroup: true}
			ch.SetDefaultBannedRights(tc.ban)
			s.rememberPeers(nil, []tg.ChatClass{ch})
			p := s.SendPermissions(peerID(&tg.PeerChannel{ChannelID: 8}))
			if p.Allows(tc.kind) || !p.Allows(model.SendAll&^tc.kind) {
				t.Fatalf("wrong independent permission: %+v", p)
			}
		})
	}
}

func TestComposerFilePermissionBeforeUpload(t *testing.T) {
	s, server := filesStore(t)
	p := s.history.peers[5]
	p.Rights.Default = model.SendPhoto
	s.history.peers[5] = p
	path := picture(t, t.TempDir(), "photo.png", 8, 8)
	msg := model.OutgoingMessage{RandomID: 1, Files: &model.OutgoingFiles{Paths: []string{path}}}
	if err := s.Send(context.Background(), 5, msg); err == nil {
		t.Fatal("forbidden photo accepted")
	}
	if server.uploads != 0 {
		t.Fatal("forbidden media uploaded")
	}
	msg.Files.Documents = true
	if err := s.Send(context.Background(), 5, msg); err != nil {
		t.Fatal("allowed document rejected:", err)
	}
	p.Rights.Default = model.SendText
	s.history.peers[5] = p
	msg.Files.Documents = false
	msg.Text = "caption"
	msg.RandomID = 2
	if err := s.Send(context.Background(), 5, msg); err != nil {
		t.Fatal("plain-text ban rejected media caption:", err)
	}
}

func TestComposerDiscussionAndBoostExemption(t *testing.T) {
	s := testStore(t)
	ch := &tg.Channel{ID: 8, Broadcast: true, HasLink: true, Photo: &tg.ChatPhotoEmpty{}}
	s.rememberPeers(nil, []tg.ChatClass{ch})
	id := peerID(&tg.PeerChannel{ChannelID: 8})
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		if _, ok := in.(*tg.ChannelsGetFullChannelRequest); !ok {
			return nil, fmt.Errorf("unexpected RPC %T", in)
		}
		full := &tg.ChannelFull{ID: 8, ChatPhoto: &tg.PhotoEmpty{}}
		full.SetLinkedChatID(9)
		full.SetBoostsUnrestrict(2)
		full.SetBoostsApplied(2)
		discussion := &tg.Channel{ID: 9, Title: "Discussion", Megagroup: true, HasLink: true, Left: true, Verified: true, Photo: &tg.ChatPhotoEmpty{}}
		discussion.SetParticipantsCount(42)
		return &tg.MessagesChatFull{FullChat: full, Chats: []tg.ChatClass{ch, discussion}}, nil
	})
	if err := s.RefreshSendPermissions(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if s.SendPermissions(id).DiscussionID != peerID(&tg.PeerChannel{ChannelID: 9}) {
		t.Fatal("linked group lost")
	}
	group, err := s.Discussion(context.Background(), id)
	if err != nil || group.Title != "Discussion" || group.Members != 42 || !group.Verified {
		t.Fatalf("discussion: %+v %v", group, err)
	}
	if !s.SendPermissions(group.ID).Allows(model.SendText) {
		t.Fatal("discussion without join_to_send needs membership")
	}
	s.history.mu.Lock()
	p := s.history.peers[id]
	p.Rights.Broadcast = false
	p.Rights.Default = model.SendPhoto
	p.Rights.Personal = model.SendVoice
	s.history.peers[id] = p
	s.history.mu.Unlock()
	rights := s.SendPermissions(id)
	if !rights.Allows(model.SendPhoto) || rights.Allows(model.SendVoice) {
		t.Fatalf("boost exemption must affect only defaults: %+v", rights)
	}
}
