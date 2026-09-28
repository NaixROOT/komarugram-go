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

// BlockedSource knows the peers the account blocked.
type BlockedSource interface {
	// Blocked reports whether the account blocked peer; the first ask
	// starts loading the list, and nothing is blocked until it is known.
	Blocked(peer int64) bool
}

// Translator translates messages, as Telegram Desktop's translate box does.
type Translator interface {
	// Translate translates message id of chat, or text when id is 0, to
	// the language to, a two-letter code.
	Translate(ctx context.Context, chat int64, id MessageID, text string, to string) (string, error)
}
