// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"fmt"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
)

// A chat's photo tells its data center; a group's online members are
// asked of Telegram, a channel's and a user's are not.
func TestChatDetails(t *testing.T) {
	s := testStore(t)
	s.rememberPeers([]tg.UserClass{
		&tg.User{ID: 5, FirstName: "Ann", Photo: &tg.UserProfilePhoto{PhotoID: 1, DCID: 4}},
		&tg.User{ID: 6, FirstName: "Bob"},
	}, []tg.ChatClass{
		&tg.Channel{ID: 8, Broadcast: true, Title: "news"},
		&tg.Channel{ID: 9, Megagroup: true, Title: "group", Photo: &tg.ChatPhoto{PhotoID: 2, DCID: 2}},
	})
	group := peerID(&tg.PeerChannel{ChannelID: 9})
	channel := peerID(&tg.PeerChannel{ChannelID: 8})
	if dc := s.ChatDetails(5).PhotoDC; dc != 4 {
		t.Errorf("user's photo DC %d", dc)
	}
	if dc := s.ChatDetails(6).PhotoDC; dc != 0 {
		t.Errorf("DC %d without a photo", dc)
	}
	if dc := s.ChatDetails(group).PhotoDC; dc != 2 {
		t.Errorf("group's photo DC %d", dc)
	}
	var asked []string
	s.history.api = composerAPI(func(in bin.Encoder) (bin.Encoder, error) {
		r, ok := in.(*tg.MessagesGetOnlinesRequest)
		if !ok {
			return nil, fmt.Errorf("unexpected %T", in)
		}
		asked = append(asked, fmt.Sprintf("%T", r.Peer))
		return &tg.ChatOnlines{Onlines: 42}, nil
	})
	for _, c := range []struct {
		chat int64
		want int
	}{{group, 42}, {channel, 0}, {5, 0}} {
		if n, err := s.Online(context.Background(), c.chat); err != nil || n != c.want {
			t.Errorf("chat %d: %d online, %v; want %d", c.chat, n, err, c.want)
		}
	}
	if len(asked) != 1 || asked[0] != "*tg.InputPeerChannel" {
		t.Errorf("asked for %v", asked)
	}
}
