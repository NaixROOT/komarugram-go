// SPDX-License-Identifier: Unlicense OR MIT

package model

// MessageNotice is a message that just came to the account, for a desktop
// notification.
type MessageNotice struct {
	Chat    Chat
	Message MessageID
	// Sender is who wrote it in a group; empty elsewhere. Text is one line,
	// as the chat list shows it.
	Sender, Text string
	// Silent is set when the sender sent it without sound.
	Silent bool
}

// NoticeSource is a store that tells of messages as they come.
type NoticeSource interface {
	// SetNotices sets the function called, on the store's goroutine, with
	// each new incoming message; nil stops it.
	SetNotices(func(MessageNotice))
}
