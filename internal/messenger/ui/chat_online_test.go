// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"testing"
	"time"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/mockstore"
	"komarugram/internal/messenger/model"
)

// A group's status tells its members online, which Telegram counts; a
// channel's does not, and the count is asked again only after a while.
func TestGroupStatusCountsOnline(t *testing.T) {
	invalidated := make(chan struct{}, 8)
	p := newChatPage(mockstore.New(time.Now(), 0), func() { invalidated <- struct{}{} })
	defer p.Close()
	l := localization.For("en")
	group := model.Chat{ID: 3, Kind: model.KindGroup, Members: 12}
	now := time.Now()
	if got := p.chatStatusOnline(group, now, l); got != "12 members" {
		t.Fatalf("before the count: %q", got)
	}
	select {
	case <-invalidated:
	case <-time.After(5 * time.Second):
		t.Fatal("the count did not redraw the page")
	}
	if got := p.chatStatusOnline(group, now, l); got != "12 members, 4 online" {
		t.Fatalf("with the count: %q", got)
	}
	asked := p.online.asked
	p.chatStatusOnline(group, now.Add(time.Second), l)
	if p.online.asked != asked {
		t.Fatal("asked again at once")
	}
	channel := model.Chat{ID: 4, Kind: model.KindChannel, Members: 48210}
	if got := p.chatStatusOnline(channel, now, l); got != "48,210 subscribers" && got != "48 210 subscribers" {
		t.Fatalf("channel: %q", got)
	}
}
