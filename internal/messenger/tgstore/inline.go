// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"komarugram/internal/messenger/model"
)

// inlineState keeps the inline bots resolved, by lowercased username, and
// the DC of web files.
type inlineState struct {
	mu    sync.Mutex
	bots  map[string]model.InlineBot
	webDC int
}

// InlineBot implements model.InlineBotSource.
func (s *Store) InlineBot(ctx context.Context, username string) (model.InlineBot, error) {
	key := strings.ToLower(username)
	s.inline.mu.Lock()
	bot, ok := s.inline.bots[key]
	s.inline.mu.Unlock()
	if ok {
		return bot, nil
	}
	api := s.api()
	if api == nil {
		return model.InlineBot{}, errors.New("offline")
	}
	res, err := s.resolveUsername(ctx, api, username)
	if errors.Is(err, model.ErrLinkNotFound) {
		return model.InlineBot{}, model.ErrNotInlineBot
	}
	if err != nil {
		return model.InlineBot{}, err
	}
	for _, raw := range res.Users {
		u, ok := raw.(*tg.User)
		if !ok || !u.Bot {
			continue
		}
		placeholder, inline := u.GetBotInlinePlaceholder()
		if !inline {
			break
		}
		bot = model.InlineBot{ID: peerID(&tg.PeerUser{UserID: u.ID}), Username: username, Placeholder: placeholder}
		s.inline.mu.Lock()
		if s.inline.bots == nil {
			s.inline.bots = map[string]model.InlineBot{}
		}
		s.inline.bots[key] = bot
		s.inline.mu.Unlock()
		return bot, nil
	}
	return model.InlineBot{}, model.ErrNotInlineBot
}

// InlineResults implements model.InlineBotSource.
func (s *Store) InlineResults(ctx context.Context, chat, bot int64, query, offset string) (model.InlineResults, error) {
	c := s.history
	c.mu.Lock()
	api, peer, botPeer := c.api, c.peers[chat], c.peers[bot]
	c.mu.Unlock()
	if api == nil {
		return model.InlineResults{}, errors.New("offline")
	}
	res, err := api.MessagesGetInlineBotResults(ctx, &tg.MessagesGetInlineBotResultsRequest{Bot: botPeer.inputUser(), Peer: peer.input(), Query: query, Offset: offset})
	if tgerr.Is(err, "BOT_RESPONSE_TIMEOUT") {
		return model.InlineResults{}, nil
	}
	if err != nil {
		return model.InlineResults{}, err
	}
	s.rememberPeers(res.Users, nil)
	dc := s.webDC(ctx, api)
	out := model.InlineResults{Gallery: res.Gallery, Next: res.NextOffset}
	refs := map[string]fileLocation{}
	for _, raw := range res.Results {
		var r model.InlineResult
		switch v := raw.(type) {
		case *tg.BotInlineResult:
			r = model.InlineResult{Kind: v.Type, Title: v.Title, Description: v.Description}
			if v.Thumb != nil {
				r.Thumb = webMedia(v.Thumb, dc, refs)
			}
			if v.Content != nil {
				r.Item.Media = model.Message{Kind: model.MessagePhoto, Media: webMedia(v.Content, dc, refs)}
				if v.Type == "gif" || strings.HasPrefix(v.Content.GetMimeType(), "video/") {
					r.Item.Media.Kind = model.MessageGIF
				}
			} else if r.Thumb != nil {
				r.Item.Media = model.Message{Kind: model.MessagePhoto, Media: r.Thumb}
			}
		case *tg.BotInlineMediaResult:
			r = model.InlineResult{Kind: v.Type, Title: v.Title, Description: v.Description}
			if d, ok := v.Document.(*tg.Document); ok {
				r.Item = documentItem(d, refs)
				r.Thumb = r.Item.Media.Media
			}
			if p, ok := v.Photo.(*tg.Photo); ok {
				c.mu.Lock()
				meta := c.rememberProfilePhoto(p, 0)
				c.mu.Unlock()
				if meta != nil {
					r.Item.Media = model.Message{Kind: model.MessagePhoto, Media: meta}
					r.Thumb = meta
				}
			}
		default:
			continue
		}
		r.Item.ID, r.Item.ResultID, r.Item.QueryID = raw.GetID(), raw.GetID(), res.QueryID
		out.Results = append(out.Results, r)
	}
	s.keepRefs(ctx, refs)
	return out, nil
}

// webDC is the DC web files are downloaded through, asked once.
func (s *Store) webDC(ctx context.Context, api *tg.Client) int {
	s.inline.mu.Lock()
	dc := s.inline.webDC
	s.inline.mu.Unlock()
	if dc != 0 {
		return dc
	}
	if cfg, err := api.HelpGetConfig(ctx); err == nil {
		dc = cfg.WebfileDCID
		s.inline.mu.Lock()
		s.inline.webDC = dc
		s.inline.mu.Unlock()
	}
	return dc
}

// webMedia is a web document of a bot, downloaded through dc; its location
// is added to refs.
func webMedia(doc tg.WebDocumentClass, dc int, refs map[string]fileLocation) *model.MessageMedia {
	id := fmt.Sprintf("inline/%x", sha256.Sum256([]byte(doc.GetURL())))
	m := &model.MessageMedia{ID: id, MIMEType: doc.GetMimeType(), Size: int64(doc.GetSize())}
	for _, attr := range doc.GetAttributes() {
		switch a := attr.(type) {
		case *tg.DocumentAttributeVideo:
			m.Width, m.Height = a.W, a.H
		case *tg.DocumentAttributeImageSize:
			m.Width, m.Height = a.W, a.H
		}
	}
	ref := fileLocation{WebURL: doc.GetURL(), DC: dc}
	if d, ok := doc.(*tg.WebDocument); ok {
		ref.WebHash = d.AccessHash
	} else {
		ref.WebNoProxy = true
	}
	refs[id] = ref
	return m
}

// StartBot implements model.BotStarter.
func (s *Store) StartBot(ctx context.Context, chat int64, param string) error {
	c := s.history
	c.mu.Lock()
	api, peer := c.api, c.peers[chat]
	c.mu.Unlock()
	if api == nil {
		return errors.New("offline")
	}
	if peer.Kind != "user" || !peer.Rights.Bot {
		return errors.New("not a bot")
	}
	res, err := api.MessagesStartBot(ctx, &tg.MessagesStartBotRequest{Bot: peer.inputUser(), Peer: peer.input(), RandomID: newRandomID(), StartParam: param})
	if err != nil {
		return err
	}
	return s.Handle(ctx, res)
}
