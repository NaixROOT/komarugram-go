// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"sync"
	"time"

	"github.com/gotd/td/tg"

	"komarugram/internal/crash"
)

// blockedPage is how many blocked peers a page of contacts.getBlocked has.
const blockedPage = 100

// blockedState is the account's blocked peers, loaded once and kept up to
// date by UpdatePeerBlocked.
type blockedState struct {
	mu      sync.Mutex
	peers   map[int64]bool
	loading bool
	retry   time.Time
}

// Blocked implements model.BlockedSource.
func (s *Store) Blocked(peer int64) bool {
	b := &s.blocked
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.peers == nil && !b.loading && time.Now().After(b.retry) {
		b.loading = true
		c := s.history
		c.mu.Lock()
		api, ctx, closing := c.api, c.ctx, c.closing
		if api != nil && !closing {
			c.wg.Go(func() {
				defer crash.Recover("blocked peers", nil)
				ctx, cancel := context.WithTimeout(ctx, time.Minute)
				defer cancel()
				s.loadBlocked(ctx, api)
			})
		} else {
			b.loading, b.retry = false, time.Now().Add(lookupRetry)
		}
		c.mu.Unlock()
	}
	return b.peers[peer]
}

// loadBlocked reads every page of contacts.getBlocked.
func (s *Store) loadBlocked(ctx context.Context, api *tg.Client) {
	peers := map[int64]bool{}
	for offset := 0; ; offset += blockedPage {
		res, err := api.ContactsGetBlocked(ctx, &tg.ContactsGetBlockedRequest{Offset: offset, Limit: blockedPage})
		if err != nil {
			s.blocked.mu.Lock()
			s.blocked.loading, s.blocked.retry = false, time.Now().Add(lookupRetry)
			s.blocked.mu.Unlock()
			return
		}
		var page []tg.PeerBlocked
		switch res := res.(type) {
		case *tg.ContactsBlocked:
			page = res.Blocked
		case *tg.ContactsBlockedSlice:
			page = res.Blocked
		}
		for _, p := range page {
			peers[peerID(p.PeerID)] = true
		}
		if len(page) < blockedPage {
			break
		}
	}
	s.blocked.mu.Lock()
	s.blocked.peers, s.blocked.loading = peers, false
	s.blocked.mu.Unlock()
	s.changed()
}

// applyBlocked follows UpdatePeerBlocked.
func (s *Store) applyBlocked(u *tg.UpdatePeerBlocked) {
	b := &s.blocked
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.peers != nil {
		b.peers[peerID(u.PeerID)] = u.Blocked
	}
}
