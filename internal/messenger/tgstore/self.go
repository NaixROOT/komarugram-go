// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"komarugram/internal/messenger/model"
)

// SelfPhotoID returns the id of the account's current profile photo, 0 when
// it has none or the profile is not loaded yet.
func (s *Store) SelfPhotoID() int64 {
	id := s.Me().ID
	c := s.history
	c.mu.Lock()
	defer c.mu.Unlock()
	if peer, ok := c.peers[id]; ok && peer.Photo != nil {
		return peer.Photo.ID
	}
	return 0
}

// SelfPhoto downloads the small (160 px) profile photo of the account, as
// the account list keeps it. It returns id 0 and no data when there is no
// photo.
func (s *Store) SelfPhoto(ctx context.Context) (int64, []byte, error) {
	me := s.Me().ID
	if me == 0 {
		return 0, nil, errors.New("tgstore: profile not loaded")
	}
	id := s.SelfPhotoID()
	if id == 0 {
		return 0, nil, nil
	}
	msg, ok := s.Avatar(me)
	if !ok {
		return 0, nil, nil
	}
	data, err := s.Media(ctx, msg)
	if err != nil {
		return 0, nil, err
	}
	return id, data, nil
}

// EditProfile implements model.ProfileEditor.
func (s *Store) EditProfile(ctx context.Context, e model.ProfileEdit) error {
	if err := e.ValidateWith(s.Premium().Limit("about_length_limit")); err != nil {
		return err
	}
	c := s.history
	c.mu.Lock()
	api := c.api
	c.mu.Unlock()
	if api == nil {
		return model.ErrProfileOffline
	}
	me := s.Me()
	if e.FirstName != me.FirstName || e.LastName != me.LastName || e.Bio != me.Bio {
		request := &tg.AccountUpdateProfileRequest{}
		request.SetFirstName(e.FirstName)
		request.SetLastName(e.LastName)
		request.SetAbout(e.Bio)
		if _, err := api.AccountUpdateProfile(ctx, request); err != nil {
			return profileError(err)
		}
		s.publish(func() { s.me.FirstName, s.me.LastName, s.me.Bio = e.FirstName, e.LastName, e.Bio })
	}
	if e.Username != me.Username {
		if _, err := api.AccountUpdateUsername(ctx, e.Username); err != nil && !tgerr.Is(err, "USERNAME_NOT_MODIFIED") {
			return profileError(err)
		}
		s.publish(func() { s.me.Username = e.Username })
	}
	return nil
}

// profileError turns what Telegram says about a profile edit into the
// model's reasons where there is one.
func profileError(err error) error {
	switch {
	case tgerr.Is(err, "FIRSTNAME_INVALID", "LASTNAME_INVALID"):
		return model.ErrNameInvalid
	case tgerr.Is(err, "ABOUT_TOO_LONG"):
		return model.ErrBioTooLong
	case tgerr.Is(err, "USERNAME_INVALID"):
		return model.ErrUsernameInvalid
	case tgerr.Is(err, "USERNAME_OCCUPIED", "USERNAME_PURCHASE_AVAILABLE"):
		return model.ErrUsernameTaken
	}
	return err
}
