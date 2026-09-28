// SPDX-License-Identifier: Unlicense OR MIT

package model

import "context"

// Keep is what the cache keeps that Telegram takes back, as AyuGram's
// saved deleted messages and edits history.
type Keep struct {
	// Deleted keeps messages others deleted, marked Deleted, but in chats
	// with bots; Edits keeps the text others' messages had before an edit.
	Deleted, Edits bool
}

// KeepStore keeps what Keep asks for, and gives back the edits kept.
type KeepStore interface {
	SetKeep(Keep)
	// MessageEdits are the versions msg had before its edits, oldest
	// first.
	MessageEdits(ctx context.Context, msg Message) ([]Message, error)
}
