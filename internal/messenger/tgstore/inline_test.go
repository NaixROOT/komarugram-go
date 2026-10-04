// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

func TestInlineBotResultsAndStart(t *testing.T) {
	s := testStore(t)
	var started *tg.MessagesStartBotRequest
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(_ context.Context, in bin.Encoder, out bin.Decoder) error {
		switch req := in.(type) {
		case *tg.ContactsResolveUsernameRequest:
			u := &tg.User{ID: 50, AccessHash: 7, Bot: true, Username: req.Username}
			if req.Username == "wiki" {
				u.SetBotInlinePlaceholder("Search")
			}
			*out.(*tg.ContactsResolvedPeer) = tg.ContactsResolvedPeer{Peer: &tg.PeerUser{UserID: 50}, Users: []tg.UserClass{u}}
		case *tg.HelpGetConfigRequest:
			out.(*tg.Config).WebfileDCID = 4
		case *tg.MessagesGetInlineBotResultsRequest:
			article := &tg.BotInlineResult{ID: "a1", Type: "article", Title: "Go", Description: "language"}
			article.SetThumb(&tg.WebDocumentNoProxy{URL: "https://x/t.jpg", MimeType: "image/jpeg"})
			*out.(*tg.MessagesBotResults) = tg.MessagesBotResults{QueryID: 99, NextOffset: "10", Results: []tg.BotInlineResultClass{article}}
		case *tg.MessagesStartBotRequest:
			started = req
			out.(*tg.UpdatesBox).Updates = &tg.Updates{}
		}
		return nil
	}))
	ctx := context.Background()
	if _, err := s.InlineBot(ctx, "plainbot"); !errors.Is(err, model.ErrNotInlineBot) {
		t.Fatalf("a bot without inline mode: %v", err)
	}
	bot, err := s.InlineBot(ctx, "wiki")
	if err != nil || bot.Placeholder != "Search" {
		t.Fatalf("bot %+v, %v", bot, err)
	}
	res, err := s.InlineResults(ctx, bot.ID, bot.ID, "go", "")
	if err != nil || len(res.Results) != 1 || res.Next != "10" {
		t.Fatalf("results %+v, %v", res, err)
	}
	r := res.Results[0]
	if r.Title != "Go" || r.Item.QueryID != 99 || r.Item.ResultID != "a1" || r.Thumb == nil || model.ItemSendKind(r.Item) != model.SendInline {
		t.Fatalf("result %+v", r)
	}
	s.history.peers[bot.ID] = peerRecord{ID: 50, Hash: 7, Kind: "user", Rights: peerRights{Bot: true}}
	if err := s.StartBot(ctx, bot.ID, "ref42"); err != nil || started == nil || started.StartParam != "ref42" {
		t.Fatalf("start: %v, %+v", err, started)
	}
}
