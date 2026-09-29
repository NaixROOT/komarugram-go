// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"sync"
	"time"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"komarugram/internal/crash"
	"komarugram/internal/messenger/model"
)

// PressButton implements model.BotStore with messages.getBotCallbackAnswer.
// A message of a thread, such as a topic, is asked of the group it lives in.
func (s *Store) PressButton(ctx context.Context, key model.MessageKey, data []byte) (model.BotAnswer, error) {
	c := s.history
	c.mu.Lock()
	api := c.api
	chat, _ := c.threadChat(key.ChatID)
	peer := c.peers[chat]
	c.mu.Unlock()
	if api == nil || peer.ID == 0 {
		return model.BotAnswer{}, errNotConnected
	}
	res, err := api.MessagesGetBotCallbackAnswer(ctx, &tg.MessagesGetBotCallbackAnswerRequest{Peer: peer.input(), MsgID: int(key.MessageID), Data: data})
	switch {
	case tgerr.Is(err, "BOT_RESPONSE_TIMEOUT"):
		return model.BotAnswer{}, model.ErrBotSilent
	case tgerr.Is(err, "PASSWORD_HASH_INVALID", "SRP_ID_INVALID"):
		return model.BotAnswer{}, model.ErrBotPassword
	case err != nil:
		return model.BotAnswer{}, err
	}
	answer := model.BotAnswer{Alert: res.Alert}
	answer.Text, _ = res.GetMessage()
	if res.HasURL {
		answer.URL, _ = res.GetURL()
	}
	return answer, nil
}

// botCommandsRetry is how long after a failed read of a bot's commands they
// are asked for again.
const botCommandsRetry = 30 * time.Second

// botState is the info of the bots read so far.
type botState struct {
	mu    sync.Mutex
	info  map[int64]model.BotInfo
	asked map[int64]time.Time
}

// BotInfo implements model.BotInfoSource with users.getFullUser, which
// carries the bot's info.
func (s *Store) BotInfo(chat int64) model.BotInfo {
	b := &s.bots
	b.mu.Lock()
	defer b.mu.Unlock()
	if info, ok := b.info[chat]; ok {
		return info
	}
	if at, ok := b.asked[chat]; ok && time.Since(at) < botCommandsRetry {
		return model.BotInfo{}
	}
	c := s.history
	c.mu.Lock()
	api, peer, ctx := c.api, c.peers[chat], c.ctx
	closing := c.closing
	c.mu.Unlock()
	if api == nil || peer.ID == 0 || peer.Kind != "user" || closing {
		return model.BotInfo{}
	}
	if b.asked == nil {
		b.asked = map[int64]time.Time{}
	}
	b.asked[chat] = time.Now()
	c.wg.Go(func() {
		defer crash.Recover("bot commands", nil)
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()
		full, err := api.UsersGetFullUser(ctx, &tg.InputUser{UserID: peer.ID, AccessHash: peer.Hash})
		if err != nil {
			return
		}
		var got model.BotInfo
		if info, ok := full.FullUser.GetBotInfo(); ok {
			list, _ := info.GetCommands()
			for _, cmd := range list {
				got.Commands = append(got.Commands, model.BotCommand{Command: cmd.Command, Description: cmd.Description})
			}
			if menu, ok := info.GetMenuButton(); ok {
				if m, ok := menu.(*tg.BotMenuButton); ok && m.URL != "" {
					got.Menu = &model.BotMenu{Text: m.Text, URL: m.URL}
				}
			}
		}
		b.mu.Lock()
		if b.info == nil {
			b.info = map[int64]model.BotInfo{}
		}
		b.info[chat] = got
		b.mu.Unlock()
		s.changed()
	})
	return model.BotInfo{}
}
