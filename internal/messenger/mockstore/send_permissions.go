package mockstore

import (
	"context"
	"errors"
	"komarugram/internal/messenger/model"
	"time"
)

// SendPermissions gives the demo representative channel and group states.
func (s *Store) SendPermissions(chat int64) model.SendPermissions {
	switch chat {
	case 4:
		return model.SendPermissions{Unavailable: true, Broadcast: true, DiscussionID: 3}
	case 8:
		return model.SendPermissions{Unavailable: true, Broadcast: true}
	case 7:
		return model.SendPermissions{Default: model.SendPhoto | model.SendVoice | model.SendGIF}
	case 10:
		return model.SendPermissions{Personal: model.SendText, Until: s.created.Add(24 * time.Hour).Unix()}
	default:
		return model.SendPermissions{}
	}
}
func (s *Store) Discussion(ctx context.Context, chat int64) (model.Chat, error) {
	if err := ctx.Err(); err != nil {
		return model.Chat{}, err
	}
	id := s.SendPermissions(chat).DiscussionID
	for _, c := range s.Chats() {
		if c.ID == id {
			return c, nil
		}
	}
	return model.Chat{}, errors.New("no discussion group")
}
