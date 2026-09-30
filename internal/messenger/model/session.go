// SPDX-License-Identifier: Unlicense OR MIT

package model

import "time"

// SessionEnd is why Telegram no longer accepts the account's session.
type SessionEnd int

const (
	// SessionAlive is a session Telegram has not ended.
	SessionAlive SessionEnd = iota
	// SessionDuplicated: the key was used over two connections at once,
	// and the server invalidated it (AUTH_KEY_DUPLICATED).
	SessionDuplicated
	// SessionRevoked: the session was ended from another device
	// (SESSION_REVOKED).
	SessionRevoked
	// SessionExpired: the session expired (SESSION_EXPIRED).
	SessionExpired
	// SessionUnregistered: the server does not know the session's key
	// (AUTH_KEY_UNREGISTERED, AUTH_KEY_INVALID).
	SessionUnregistered
	// SessionDeleted: the account was deleted (USER_DEACTIVATED).
	SessionDeleted
	// SessionBanned: the account was banned (USER_DEACTIVATED_BAN).
	SessionBanned
)

// SessionSource is a Store that knows whether Telegram still accepts the
// account's session.
type SessionSource interface {
	// SessionEnded reports why Telegram ended the session on its side:
	// nothing loads any more, and only signing in again brings the account
	// back. It is SessionAlive while the session works.
	SessionEnded() SessionEnd
}

// ConnectionSource is a Store that knows whether the account's connection
// stopped for a reason other than its session ending, such as its local
// data not opening or the account being connected elsewhere.
type ConnectionSource interface {
	// ConnectionFailed is what stopped the connection, nil while it runs or
	// is being made. Nothing loads until Reconnect; what is saved can still
	// be read.
	ConnectionFailed() error
	// Reconnect starts the connection again after it failed.
	Reconnect()
}

// Freeze is Telegram's freezing of an account for breaking its terms: the
// account can read but not send or change anything, and is deleted at
// Until unless an appeal succeeds.
type Freeze struct {
	Since, Until time.Time
	// AppealURL is where to appeal, a link to @SpamBot.
	AppealURL string
}

// Frozen reports whether the account is frozen.
func (f Freeze) Frozen() bool { return !f.Since.IsZero() }

// FreezeSource is a Store that knows whether the account is frozen.
type FreezeSource interface {
	Freeze() Freeze
}
