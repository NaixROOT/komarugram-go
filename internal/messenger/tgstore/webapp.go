// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"time"

	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

// webPlatform is the name this client gives Mini Apps as its platform. Apps
// tell desktop from mobile by it, and know only Telegram's own names.
const webPlatform = "tdesktop"

// bot returns the bot as an input user, and the chat's peer.
func (s *Store) webViewPeers(req model.WebViewRequest) (api *tg.Client, bot tg.InputUserClass, chat tg.InputPeerClass, err error) {
	c := s.history
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.api == nil {
		return nil, nil, nil, errNotConnected
	}
	b := c.peers[req.Bot]
	if b.ID == 0 || b.Kind != "user" {
		return nil, nil, nil, errors.New("the bot is not known")
	}
	bot = &tg.InputUser{UserID: b.ID, AccessHash: b.Hash}
	real, _ := c.threadChat(req.Chat)
	if p := c.peers[real]; p.ID != 0 {
		chat = p.input()
	}
	return c.api, bot, chat, nil
}

// RequestWebView implements model.WebViewStore: a keyboard's SimpleWebView
// button asks messages.requestSimpleWebView, the others
// messages.requestWebView, from the bot's menu or from the chat.
func (s *Store) RequestWebView(ctx context.Context, req model.WebViewRequest) (model.WebView, error) {
	api, bot, chat, err := s.webViewPeers(req)
	if err != nil {
		return model.WebView{}, err
	}
	theme := tg.DataJSON{Data: req.Theme}
	if req.Kind == model.WebViewSimple {
		r := &tg.MessagesRequestSimpleWebViewRequest{Bot: bot, ThemeParams: theme, Platform: webPlatform}
		r.SetURL(req.URL)
		if req.StartParam != "" {
			r.SetStartParam(req.StartParam)
		}
		res, err := api.MessagesRequestSimpleWebView(ctx, r)
		if err != nil {
			return model.WebView{}, err
		}
		return model.WebView{URL: res.URL}, nil
	}
	if chat == nil {
		chat = &tg.InputPeerEmpty{}
	}
	r := &tg.MessagesRequestWebViewRequest{Peer: chat, Bot: bot, ThemeParams: theme, Platform: webPlatform, FromBotMenu: req.Kind == model.WebViewMenu}
	if req.URL != "" {
		r.SetURL(req.URL)
	}
	if req.StartParam != "" {
		r.SetStartParam(req.StartParam)
	}
	res, err := api.MessagesRequestWebView(ctx, r)
	if err != nil {
		return model.WebView{}, err
	}
	return model.WebView{URL: res.URL, QueryID: res.QueryID}, nil
}

// ProlongWebView implements model.WebViewStore.
func (s *Store) ProlongWebView(ctx context.Context, req model.WebViewRequest, queryID int64) error {
	api, bot, chat, err := s.webViewPeers(req)
	if err != nil {
		return err
	}
	if chat == nil {
		chat = &tg.InputPeerEmpty{}
	}
	_, err = api.MessagesProlongWebView(ctx, &tg.MessagesProlongWebViewRequest{Peer: chat, Bot: bot, QueryID: queryID})
	return err
}

// SendWebViewData implements model.WebViewStore with
// messages.sendWebViewData; the message it makes comes back in the updates.
func (s *Store) SendWebViewData(ctx context.Context, botID int64, buttonText, data string) error {
	api, bot, _, err := s.webViewPeers(model.WebViewRequest{Bot: botID})
	if err != nil {
		return err
	}
	updates, err := api.MessagesSendWebViewData(ctx, &tg.MessagesSendWebViewDataRequest{Bot: bot, RandomID: newRandomID(), ButtonText: buttonText, Data: data})
	if err != nil {
		return err
	}
	return s.Handle(ctx, updates)
}

// newRandomID is a random_id for a request that sends something.
func newRandomID() int64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return time.Now().UnixNano()
	}
	if v := int64(binary.LittleEndian.Uint64(b[:])); v != 0 {
		return v
	}
	return 1
}
