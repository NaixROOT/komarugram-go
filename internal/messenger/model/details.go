// SPDX-License-Identifier: Unlicense OR MIT

package model

import (
	"context"
	"time"
)

// ChatDetails are facts about a chat that its info shows, as materialgram's
// profile does.
type ChatDetails struct {
	// PhotoDC is the data center that keeps the chat's photo, 0 without a
	// photo.
	PhotoDC int
}

// ChatDetailer tells the details of chats.
type ChatDetailer interface {
	// ChatDetails are what is known of chat without asking Telegram.
	ChatDetails(chat int64) ChatDetails
	// Online is how many members of a group are online now.
	Online(ctx context.Context, chat int64) (int, error)
}

// Registration tells how RegisteredAround's date relates to the account's
// registration.
type Registration int

const (
	// RegisteredAbout: the account was registered around the date.
	RegisteredAbout Registration = iota
	// RegisteredBefore: before the date, older than every known account.
	RegisteredBefore
	// RegisteredAfter: after the date, newer than every known account.
	RegisteredAfter
)

// registrations are user ids with the dates their accounts were
// registered. Telegram gives ids in order, so an id between two of them
// was registered between their dates.
var registrations = []struct {
	id   int64
	date int64
}{
	{1000000, 1380326400},
	{2768409, 1383264000},
	{7679610, 1388448000},
	{11538514, 1391212000},
	{6386727079, 1691696880},
	{6429580803, 1692082680},
	{6813121418, 1698489600},
	{6865576492, 1699052400},
	{6925870357, 1701192327},
	{7000000000, 1711889200},
}

// RegisteredAround estimates when the account of user was registered, as
// materialgram does: between the dates of the known ids around it.
func RegisteredAround(user int64) (time.Time, Registration) {
	first, last := registrations[0], registrations[len(registrations)-1]
	switch {
	case user < first.id:
		return time.Unix(first.date, 0), RegisteredBefore
	case user > last.id:
		return time.Unix(last.date, 0), RegisteredAfter
	}
	for i := 1; i < len(registrations); i++ {
		a, b := registrations[i-1], registrations[i]
		if user <= b.id {
			t := float64(user-a.id) / float64(b.id-a.id)
			return time.Unix(a.date+int64(t*float64(b.date-a.date)), 0), RegisteredAbout
		}
	}
	return time.Unix(last.date, 0), RegisteredAfter
}

// DataCenterName is where Telegram's data center dc is.
func DataCenterName(dc int) string {
	switch dc {
	case 1, 3:
		return "Miami"
	case 2, 4:
		return "Amsterdam"
	case 5:
		return "Singapore"
	}
	return ""
}
