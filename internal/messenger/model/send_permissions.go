package model

import (
	"context"
	"time"
)

// SendKind identifies independently restricted Telegram composer actions.
type SendKind uint32

const (
	SendText SendKind = 1 << iota
	SendPhoto
	SendVideo
	SendMusic
	SendFile
	SendVoice
	SendRoundVideo
	SendSticker
	SendGIF
	SendInline
	SendPoll
	SendGame
	SendLinkPreview
)
const SendAll = SendText | SendPhoto | SendVideo | SendMusic | SendFile | SendVoice | SendRoundVideo | SendSticker | SendGIF | SendInline | SendPoll | SendGame | SendLinkPreview
const SendAttachments = SendPhoto | SendVideo | SendMusic | SendFile | SendGIF

// SendPermissions keeps default and personal bans separate, for explanations.
// Unavailable also covers unknown peers and old caches without rights.
type SendPermissions struct {
	Unavailable, Broadcast bool
	Default, Personal      SendKind
	Until                  int64
	DiscussionID           int64
}

func (p SendPermissions) Allows(k SendKind) bool {
	return !p.Unavailable && (p.Default|p.Personal)&k == 0
}
func (p SendPermissions) Any(k SendKind) bool {
	return !p.Unavailable && k & ^(p.Default|p.Personal) != 0
}
func (p SendPermissions) Expiry(k SendKind) time.Time {
	if p.Default&k != 0 || p.Personal&k == 0 || p.Until <= 0 || p.Until == 2147483647 {
		return time.Time{}
	}
	return time.Unix(p.Until, 0)
}
func (k SendKind) Key() string {
	switch k {
	case SendPhoto:
		return "photos"
	case SendVideo:
		return "videos"
	case SendMusic:
		return "music"
	case SendFile:
		return "files"
	case SendVoice:
		return "voice_messages"
	case SendRoundVideo:
		return "video_messages"
	case SendSticker:
		return "stickers"
	case SendGIF:
		return "gifs"
	case SendInline:
		return "inline"
	case SendPoll:
		return "polls"
	default:
		return "message"
	}
}

type SendRestriction struct {
	Kind        SendKind
	Permissions SendPermissions
}

func (e *SendRestriction) Error() string {
	return "sending " + e.Kind.Key() + " is not allowed in this chat"
}
func (p SendPermissions) Check(k SendKind) error {
	if p.Allows(k) {
		return nil
	}
	return &SendRestriction{Kind: k, Permissions: p}
}
func PickerSendKind(tab PickerTab) SendKind {
	switch tab {
	case PickerStickers:
		return SendSticker
	case PickerGIF:
		return SendGIF
	default:
		return SendText
	}
}
func ItemSendKind(item PickerItem) SendKind {
	switch {
	case item.Media.Kind == MessageGIF:
		return SendGIF
	case item.ResultID != "":
		// A result of an inline bot, which send_inline allows.
		return SendInline
	}
	return SendSticker
}

type SendPermissionsSource interface{ SendPermissions(int64) SendPermissions }

// DiscussionSource resolves channelFull.linked_chat_id without joining the group.
type DiscussionSource interface {
	Discussion(context.Context, int64) (Chat, error)
}

// SendPermissionsRefresher refreshes full peer data while a composer is open.
type SendPermissionsRefresher interface {
	RefreshSendPermissions(context.Context, int64) error
}
