// SPDX-License-Identifier: Unlicense OR MIT

package model

import (
	"testing"
	"time"
)

func TestRegisteredAround(t *testing.T) {
	for _, c := range []struct {
		user int64
		want string
		how  Registration
	}{
		{500, "2013-09", RegisteredBefore},
		{1000000, "2013-09", RegisteredAbout},
		// Halfway between two known ids is halfway between their dates.
		{(6813121418 + 6865576492) / 2, "2023-10", RegisteredAbout},
		{6925870357, "2023-11", RegisteredAbout},
		{8000000000, "2024-03", RegisteredAfter},
	} {
		at, how := RegisteredAround(c.user)
		if got := at.UTC().Format("2006-01"); got != c.want || how != c.how {
			t.Errorf("user %d: %s %d, want %s %d", c.user, got, how, c.want, c.how)
		}
	}
	a, _ := RegisteredAround(6813121418)
	b, _ := RegisteredAround(6865576492)
	mid, _ := RegisteredAround((6813121418 + 6865576492) / 2)
	if d := mid.Sub(a) - b.Sub(mid); d > time.Minute || d < -time.Minute {
		t.Errorf("not halfway: %v", d)
	}
}
