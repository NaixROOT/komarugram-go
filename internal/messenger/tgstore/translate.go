// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"

	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

// Translate implements model.Translator with messages.translateText: by
// the message, which Telegram reads itself, or by the text.
func (s *Store) Translate(ctx context.Context, chat int64, id model.MessageID, text string, to string) (string, error) {
	c := s.history
	c.mu.Lock()
	real, _ := c.threadChat(chat)
	api, peer := c.api, c.peers[real]
	c.mu.Unlock()
	if api == nil {
		return "", errNotConnected
	}
	req := &tg.MessagesTranslateTextRequest{ToLang: to}
	if id > 0 && peer.ID != 0 {
		req.SetPeer(peer.input())
		req.SetID([]int{int(id)})
	} else {
		req.SetText([]tg.TextWithEntities{{Text: text}})
	}
	res, err := api.MessagesTranslateText(ctx, req)
	if err != nil {
		return "", err
	}
	if len(res.Result) == 0 {
		return "", errors.New("nothing translated")
	}
	return res.Result[0].Text, nil
}
