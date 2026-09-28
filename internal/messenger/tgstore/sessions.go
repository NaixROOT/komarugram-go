// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"time"

	"github.com/gotd/td/tg"

	"komarugram/internal/messenger/model"
)

// Sessions implements model.SessionsSource with account.getAuthorizations.
func (s *Store) Sessions(ctx context.Context) ([]model.Session, error) {
	s.history.mu.Lock()
	api := s.history.api
	s.history.mu.Unlock()
	if api == nil {
		return nil, errNotConnected
	}
	res, err := api.AccountGetAuthorizations(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]model.Session, 0, len(res.Authorizations))
	for _, a := range res.Authorizations {
		out = append(out, convertSession(a))
	}
	return out, nil
}

func convertSession(a tg.Authorization) model.Session {
	s := model.Session{
		Current:    a.Current,
		Incomplete: a.PasswordPending,
		Official:   a.OfficialApp,
		Hash:       a.Hash,
		APIID:      a.APIID,
		Device:     a.DeviceModel,
		Platform:   a.Platform,
		System:     a.SystemVersion,
		App:        model.SessionApp(a.APIID, a.AppName, a.AppVersion),
		IP:         a.IP,
		Country:    a.Country,
		Created:    time.Unix(int64(a.DateCreated), 0),
		Active:     time.Unix(int64(a.DateActive), 0),
	}
	if a.DateActive == 0 {
		s.Active = s.Created
	}
	return s
}
