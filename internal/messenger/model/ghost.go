// SPDX-License-Identifier: Unlicense OR MIT

package model

// Ghost is what the account tells others of itself, as AyuGram's Ghost
// Mode: the zero value tells nothing, which is this client's way.
type Ghost struct {
	// SendRead marks messages read as they are shown; SendOnline shows the
	// account online while a window is active; SendTyping tells a chat the
	// account types in it.
	SendRead, SendOnline, SendTyping bool
	// ReadOnInteract marks a chat read, without SendRead, when the account
	// sends to it or reacts in it.
	ReadOnInteract bool
}

// GhostStore sends what Ghost allows.
type GhostStore interface {
	SetGhost(Ghost)
	Ghost() Ghost
	// MarkRead marks chat read up to message id, when Ghost.SendRead, or
	// always when asked, as by the menu's Read Message.
	MarkRead(chat int64, id MessageID, asked bool)
	// SetOnline tells whether a window of the account is active.
	SetOnline(bool)
	// Typing tells that the account types in chat.
	Typing(chat int64)
}
