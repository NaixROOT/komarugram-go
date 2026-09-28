// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"time"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"komarugram/internal/crash"
)

// Telegram pushes the updates of a channel only to its members. A channel the
// account has not joined — opened from search or a link — stays as it was
// loaded unless it is asked: while one is on screen, its difference is polled
// the way Telegram Desktop does (Api::Updates::addActiveChat): a second after
// it is shown, and then as often as the server's timeout says.
//
// Channels the account is in are left to the updates manager, which keeps
// their pts; asking for their difference here would make it skip updates.
var (
	// pollFirst is the wait before the first question
	// (kWaitForChannelGetDifference).
	pollFirst = time.Second
	// pollDefault is the wait when the server gives no timeout.
	pollDefault = time.Second
	// pollRetry is the wait after an error; it doubles up to pollRetryMax.
	pollRetry    = 5 * time.Second
	pollRetryMax = 5 * time.Minute
)

// pollLimit is the most messages asked for at once
// (kChannelGetDifferenceLimit).
const pollLimit = 100

// channelPolls is what polls the channels on screen.
type channelPolls struct {
	// watched is the chat each viewer shows; polls, the running polls by
	// chat.
	watched map[any]int64
	polls   map[int64]context.CancelFunc
}

// WatchChat implements model.ChatWatcher: viewer, such as a window's history,
// now shows chat; 0 means it shows none. It is called on every frame, so the
// poll starts as soon as the channel is known to be one the account left.
func (s *Store) WatchChat(viewer any, chat int64) {
	c := s.history
	c.mu.Lock()
	defer c.mu.Unlock()
	w := &c.watch
	if w.watched == nil {
		w.watched, w.polls = map[any]int64{}, map[int64]context.CancelFunc{}
	}
	old := w.watched[viewer]
	if isThread(chat) {
		// A thread is kept up to date through its discussion group, 0
		// until it is found.
		chat, _ = c.threadChat(chat)
	}
	if chat == 0 {
		delete(w.watched, viewer)
	} else {
		w.watched[viewer] = chat
	}
	if old != 0 && old != chat && !w.shown(old) {
		if cancel := w.polls[old]; cancel != nil {
			cancel()
			delete(w.polls, old)
		}
	}
	if chat == 0 || w.polls[chat] != nil || c.closing {
		return
	}
	peer, ok := c.peers[chat]
	if !ok || peer.Kind != "channel" || !peer.Rights.Left {
		return
	}
	ctx, cancel := context.WithCancel(c.ctx)
	w.polls[chat] = cancel
	c.wg.Go(func() {
		defer crash.Recover("channel poll", nil)
		s.pollChannel(ctx, chat)
	})
}

// shown reports whether a viewer shows chat.
func (w *channelPolls) shown(chat int64) bool {
	for _, id := range w.watched {
		if id == chat {
			return true
		}
	}
	return false
}

// stopPolls ends every poll, when the store closes.
func (w *channelPolls) stopPolls() {
	for chat, cancel := range w.polls {
		cancel()
		delete(w.polls, chat)
	}
}

// pollChannel asks for the difference of chat, a channel the account is not
// in, until ctx ends.
func (s *Store) pollChannel(ctx context.Context, chat int64) {
	wait, retry := pollFirst, pollRetry
	pts := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
		}
		c := s.history
		c.mu.Lock()
		api, peer := c.api, c.peers[chat]
		c.mu.Unlock()
		if api == nil {
			wait = pollRetry
			continue
		}
		channel := &tg.InputChannel{ChannelID: peer.ID, AccessHash: peer.Hash}
		var err error
		if pts == 0 {
			pts, err = channelPts(ctx, api, channel)
		} else {
			var next int
			next, wait, err = s.channelDifference(ctx, api, chat, channel, pts)
			if err == nil {
				pts = next
			}
		}
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			if d, ok := tgerr.AsFloodWait(err); ok {
				wait = d
			} else {
				wait, retry = retry, min(2*retry, pollRetryMax)
			}
			continue
		}
		retry = pollRetry
	}
}

// channelPts is where the channel's updates stand now.
func channelPts(ctx context.Context, api *tg.Client, channel *tg.InputChannel) (int, error) {
	full, err := api.ChannelsGetFullChannel(ctx, channel)
	if err != nil {
		return 0, err
	}
	f, ok := full.FullChat.(*tg.ChannelFull)
	if !ok {
		return 0, errors.New("not a channel")
	}
	return f.Pts, nil
}

// channelDifference asks for what happened in the channel since pts and
// applies it. It returns the new pts and how long to wait before asking
// again; no wait when the answer was not the whole difference.
func (s *Store) channelDifference(ctx context.Context, api *tg.Client, chat int64, channel *tg.InputChannel, pts int) (int, time.Duration, error) {
	diff, err := api.UpdatesGetChannelDifference(ctx, &tg.UpdatesGetChannelDifferenceRequest{
		Channel: channel,
		Filter:  &tg.ChannelMessagesFilterEmpty{},
		Pts:     pts,
		Limit:   pollLimit,
	})
	if err != nil {
		return pts, 0, err
	}
	timeout := func(seconds int, ok bool) time.Duration {
		if !ok || seconds <= 0 {
			return pollDefault
		}
		return time.Duration(seconds) * time.Second
	}
	switch d := diff.(type) {
	case *tg.UpdatesChannelDifferenceEmpty:
		return d.Pts, timeout(d.GetTimeout()), nil
	case *tg.UpdatesChannelDifference:
		updates := make([]tg.UpdateClass, 0, len(d.NewMessages)+len(d.OtherUpdates))
		for _, m := range d.NewMessages {
			updates = append(updates, &tg.UpdateNewChannelMessage{Message: m})
		}
		updates = append(updates, d.OtherUpdates...)
		if err := s.Handle(ctx, &tg.Updates{Updates: updates, Users: d.Users, Chats: d.Chats}); err != nil {
			return pts, 0, err
		}
		if !d.Final {
			return d.Pts, 0, nil
		}
		return d.Pts, timeout(d.GetTimeout()), nil
	case *tg.UpdatesChannelDifferenceTooLong:
		// Too much happened to replay: take the history again from the
		// server, as after a gap.
		s.rememberPeers(d.Users, d.Chats)
		s.Reload(chat)
		next := pts
		if dialog, ok := d.Dialog.(*tg.Dialog); ok {
			if p, ok := dialog.GetPts(); ok {
				next = p
			}
		}
		return next, timeout(d.GetTimeout()), nil
	}
	return pts, pollDefault, nil
}
