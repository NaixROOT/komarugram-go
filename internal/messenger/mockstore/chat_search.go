// SPDX-License-Identifier: Unlicense OR MIT

package mockstore

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"komarugram/internal/messenger/model"
)

// SearchChat finds the demo chat's messages that contain text, newest
// first.
func (s *Store) SearchChat(ctx context.Context, chat int64, text, next string, limit int) (model.ChatSearchPage, error) {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return model.ChatSearchPage{}, nil
	}
	var found []model.Message
	for _, m := range s.History(chat).Messages {
		if strings.Contains(strings.ToLower(m.Text), text) {
			found = append(found, m)
		}
	}
	slices.Reverse(found)
	offset, _ := strconv.Atoi(next)
	page := model.ChatSearchPage{Count: len(found)}
	if offset < len(found) {
		page.Messages = found[offset:min(offset+limit, len(found))]
	}
	if offset+limit < len(found) {
		page.Next = strconv.Itoa(offset + limit)
	}
	return page, ctx.Err()
}
