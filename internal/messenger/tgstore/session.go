// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"time"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"komarugram/internal/messenger/model"
)

// freezeRefresh is how often a FROZEN_METHOD_INVALID may read the client
// configuration again, as Telegram Desktop does.
const freezeRefresh = time.Hour

// SessionEnd tells whether err is Telegram ending the session for good, so
// that connecting again cannot help, and why. These are the errors
// https://core.telegram.org/api/errors lists for an authorization that is
// gone: AUTH_KEY_PERM_EMPTY is not one, as it is about a temporary key.
func SessionEnd(err error) model.SessionEnd {
	switch {
	case err == nil:
		return model.SessionAlive
	case tgerr.Is(err, "AUTH_KEY_DUPLICATED"):
		return model.SessionDuplicated
	case tgerr.Is(err, "SESSION_REVOKED"):
		return model.SessionRevoked
	case tgerr.Is(err, "SESSION_EXPIRED"):
		return model.SessionExpired
	case tgerr.Is(err, "AUTH_KEY_UNREGISTERED", "AUTH_KEY_INVALID"):
		return model.SessionUnregistered
	case tgerr.Is(err, "USER_DEACTIVATED_BAN"):
		return model.SessionBanned
	case tgerr.Is(err, "USER_DEACTIVATED"):
		return model.SessionDeleted
	}
	return model.SessionAlive
}

// EndSession records why Telegram ended the session: see
// model.SessionSource.
func (s *Store) EndSession(why model.SessionEnd) {
	s.publish(func() { s.sessionEnded = why })
}

// SessionEnded implements model.SessionSource.
func (s *Store) SessionEnded() model.SessionEnd {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionEnded
}

// FailConnection records that the account's connection stopped with err and
// does not come back on its own: see model.ConnectionSource.
func (s *Store) FailConnection(err error) {
	s.publish(func() { s.connectionErr = err })
}

// ConnectionFailed implements model.ConnectionSource.
func (s *Store) ConnectionFailed() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connectionErr
}

// Reconnect implements model.ConnectionSource.
func (s *Store) Reconnect() {
	s.mu.Lock()
	failed := s.connectionErr != nil
	s.connectionErr = nil
	s.mu.Unlock()
	if !failed {
		return
	}
	select {
	case s.reconnect <- struct{}{}:
	default:
	}
	s.changed()
}

// Reconnects is sent to when Reconnect asks for the connection again.
func (s *Store) Reconnects() <-chan struct{} { return s.reconnect }

// Freeze implements model.FreezeSource.
func (s *Store) Freeze() model.Freeze {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.freeze
}

// Middleware reads the client configuration again when Telegram refuses a
// request with FROZEN_METHOD_INVALID and the account is not known to be
// frozen: the account was frozen since the configuration was read.
func (s *Store) Middleware() telegram.Middleware {
	return telegram.MiddlewareFunc(func(next tg.Invoker) telegram.InvokeFunc {
		return func(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
			err := next.Invoke(ctx, input, output)
			if tgerr.Is(err, "FROZEN_METHOD_INVALID") {
				go s.frozenMethod()
			}
			return err
		}
	})
}

// frozenMethod reads the configuration again, at most once in
// freezeRefresh. It runs on a goroutine of its own, as the middleware may be
// called with the conversation locked.
func (s *Store) frozenMethod() {
	s.history.mu.Lock()
	api, ctx := s.history.api, s.history.ctx
	s.history.mu.Unlock()
	if api == nil || ctx == nil {
		return
	}
	s.mu.Lock()
	now := time.Now()
	due := !s.freeze.Frozen() && (s.freezeChecked.IsZero() || now.Sub(s.freezeChecked) > freezeRefresh)
	if due {
		s.freezeChecked = now
	}
	s.mu.Unlock()
	if due {
		s.loadAppConfig(ctx, api, s.Me().Premium)
	}
}

// freezeFromConfig reads the freeze from the client configuration.
func freezeFromConfig(config map[string]any) model.Freeze {
	unix := func(key string) time.Time {
		if v, ok := config[key].(float64); ok && v > 0 {
			return time.Unix(int64(v), 0)
		}
		return time.Time{}
	}
	url, _ := config["freeze_appeal_url"].(string)
	return model.Freeze{Since: unix("freeze_since_date"), Until: unix("freeze_until_date"), AppealURL: url}
}

// Remember shows p, what is known of the account without asking Telegram,
// until Load reads the profile.
func (s *Store) Remember(p model.Profile) {
	s.publish(func() {
		if s.me.ID == 0 {
			s.me = p
		}
	})
}

var (
	_ model.SessionSource    = (*Store)(nil)
	_ model.ConnectionSource = (*Store)(nil)
	_ model.FreezeSource     = (*Store)(nil)
)
