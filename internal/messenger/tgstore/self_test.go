// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"testing"

	"github.com/gotd/td/tgerr"

	"komarugram/internal/messenger/model"
)

func TestProfileErrors(t *testing.T) {
	for code, want := range map[string]error{
		"USERNAME_OCCUPIED": model.ErrUsernameTaken,
		"USERNAME_INVALID":  model.ErrUsernameInvalid,
		"FIRSTNAME_INVALID": model.ErrNameInvalid,
		"ABOUT_TOO_LONG":    model.ErrBioTooLong,
	} {
		if got := profileError(tgerr.New(400, code)); !errors.Is(got, want) {
			t.Errorf("%s: %v, want %v", code, got, want)
		}
	}
	other := tgerr.New(420, "FLOOD_WAIT_10")
	if got := profileError(other); got != other {
		t.Errorf("an unknown error was replaced: %v", got)
	}
}

func TestEditProfileOffline(t *testing.T) {
	s := New(nil)
	if err := s.EditProfile(context.Background(), model.ProfileEdit{FirstName: "A"}); !errors.Is(err, model.ErrProfileOffline) {
		t.Fatalf("edit without a connection: %v", err)
	}
	if err := s.EditProfile(context.Background(), model.ProfileEdit{}); !errors.Is(err, model.ErrNameInvalid) {
		t.Fatalf("invalid edit: %v", err)
	}
}
