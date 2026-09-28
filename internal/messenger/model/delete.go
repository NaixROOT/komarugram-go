package model

import "context"

// MessageDeleter removes explicit message IDs. Channels and supergroups only
// support deletion for everyone; private chats also support local deletion.
type MessageDeleter interface {
	DeleteMessages(context.Context, int64, []MessageID, bool) error
}
