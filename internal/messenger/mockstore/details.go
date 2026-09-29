// SPDX-License-Identifier: Unlicense OR MIT

package mockstore

import (
	"context"

	"komarugram/internal/messenger/model"
)

// usernames are the public names of the demo's public chats.
var usernames = map[int64]string{2: "anna_smirnova", 4: "golang_news", 8: "material_design"}

// Username implements model.UsernameSource.
func (s *Store) Username(chat int64) string { return usernames[chat] }

// ChatDetails implements model.ChatDetailer: the demo's photos are spread
// over Telegram's data centers.
func (s *Store) ChatDetails(chat int64) model.ChatDetails {
	if chat <= 0 || chat == 1 {
		return model.ChatDetails{}
	}
	return model.ChatDetails{PhotoDC: int(chat%5) + 1}
}

// Online implements model.ChatDetailer: a third of a group's members.
func (s *Store) Online(_ context.Context, chat int64) (int, error) {
	for _, c := range s.chats {
		if c.ID == chat && c.Kind == model.KindGroup {
			return c.Members / 3, nil
		}
	}
	return 0, nil
}
