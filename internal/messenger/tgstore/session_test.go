// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"fmt"
	"testing"
	"time"

	"github.com/gotd/td/tgerr"

	"komarugram/internal/messenger/model"
)

func TestSessionEnd(t *testing.T) {
	for typ, want := range map[string]model.SessionEnd{
		"AUTH_KEY_DUPLICATED":   model.SessionDuplicated,
		"SESSION_REVOKED":       model.SessionRevoked,
		"SESSION_EXPIRED":       model.SessionExpired,
		"AUTH_KEY_UNREGISTERED": model.SessionUnregistered,
		"AUTH_KEY_INVALID":      model.SessionUnregistered,
		"USER_DEACTIVATED":      model.SessionDeleted,
		"USER_DEACTIVATED_BAN":  model.SessionBanned,
		"AUTH_KEY_PERM_EMPTY":   model.SessionAlive,
		"FLOOD_WAIT":            model.SessionAlive,
	} {
		err := fmt.Errorf("load: %w", tgerr.New(401, typ))
		if got := SessionEnd(err); got != want {
			t.Errorf("%s: %v, want %v", typ, got, want)
		}
	}
	if SessionEnd(nil) != model.SessionAlive {
		t.Error("no error ends the session")
	}
}

func TestFreezeFromConfig(t *testing.T) {
	f := freezeFromConfig(map[string]any{
		"freeze_since_date": float64(1788000000),
		"freeze_until_date": float64(1790000000),
		"freeze_appeal_url": "https://t.me/SpamBot",
	})
	if !f.Frozen() || !f.Until.Equal(time.Unix(1790000000, 0)) || f.AppealURL != "https://t.me/SpamBot" {
		t.Errorf("frozen config read as %+v", f)
	}
	if freezeFromConfig(map[string]any{"premium_purchase_blocked": false}).Frozen() {
		t.Error("an account without freeze keys is frozen")
	}
}
